package cloud

import (
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/resty.v1"
)

const(
	AWS = "aws"
	Azure = "azure"
	GCP = "gcp"
	Unknown = "unknown"
)

func MetaFetch() cty.Value {
	meta := make(map[string]cty.Value)
	tags := make(map[string]cty.Value)

	meta["provider"] = cty.StringVal(Unknown)

	// Return quickly when no metadata endpoint is available
	if !MetaIsPresent() {
		return cty.ObjectVal(meta)
	}

	// AWS
	sess := session.Must(session.NewSession())
	awsMeta := ec2metadata.New(sess)

	if awsMeta.Available() {
		region, err := awsMeta.Region()

		if err == nil {
			sess.Config.Region = aws.String(region)
			ec2client := ec2.New(sess)

			instance, err := awsMeta.GetInstanceIdentityDocument()
			if err == nil {
				res, err := ec2client.DescribeInstances(&ec2.DescribeInstancesInput{
					InstanceIds: aws.StringSlice([]string{instance.InstanceID}),
				})
				if err == nil {
					for _, tag := range res.Reservations[0].Instances[0].Tags {
						tags[*tag.Key] = cty.StringVal(*tag.Value)
					}

					meta["provider"] = cty.StringVal(AWS)
					meta["tags"] = cty.MapVal(tags)

					return cty.ObjectVal(meta)
				}
			}
		}
	}

	// Azure
	az := resty.New()
	az.SetTimeout(time.Duration(5) * time.Second)

	azres, err := az.R().
		SetHeader("Metadata", "true").
		SetResult(&AzureMeta{}).
		Get("http://169.254.169.254/metadata/instance?api-version=2020-09-01")

	if err == nil && azres.StatusCode() != 404 {
		if azres.Result().(*AzureMeta).Compute.Tags != "" {
			for _, tag := range strings.Split(azres.Result().(*AzureMeta).Compute.Tags, ";" ) {
				split := strings.Split(tag, ":")
				tags[split[0]] = cty.StringVal(split[1])
			}

			meta["tags"] = cty.MapVal(tags)
		}

		meta["provider"] = cty.StringVal(Azure)

		return cty.ObjectVal(meta)
	}

	// GCP
	gcp := metadata.NewClient(&http.Client{
		Timeout: time.Second * 5,
	})

	attrs, err := gcp.InstanceAttributes()
	if err == nil {
		for _, attr := range attrs {
			val, err := gcp.InstanceAttributeValue(attr)
			if err == nil {
				tags[attr] = cty.StringVal(val)
			}
		}

		meta["provider"] = cty.StringVal(GCP)
		meta["tags"] = cty.MapVal(tags)

		return cty.ObjectVal(meta)
	}

	return cty.ObjectVal(meta)
}

func MetaIsPresent() bool {
	check := resty.New()

	dialer := &net.Dialer{
		Timeout:   100 * time.Millisecond,
		KeepAlive: 30 * time.Second,
	}

	check.SetTransport(&http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	})

	check.SetCloseConnection(true)

	_, err := check.R().Get("http://169.254.169.254/")
	if _, ok := err.(net.Error); ok {
		return false
	}

	return true
}