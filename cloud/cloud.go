package cloud

import (
	"net/http"
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

type AzureMeta struct {
	Compute *AzureMetaCompute `json:"compute"`
}

type AzureMetaCompute struct {
	Tags string `json:"tags"`
}

func MetaFetch() cty.Value {
	meta := make(map[string]cty.Value)
	tags := make(map[string]cty.Value)

	meta["provider"] = cty.StringVal(Unknown)

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

	if err == nil {
		for _, tag := range strings.Split(azres.Result().(*AzureMeta).Compute.Tags, ";" ) {
			split := strings.Split(tag, ":")
			tags[split[0]] = cty.StringVal(split[1])
		}

		meta["provider"] = cty.StringVal(Azure)
		meta["tags"] = cty.MapVal(tags)

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

func MetaUnknown() cty.Value {
	return cty.ObjectVal(map[string]cty.Value{"provider": cty.StringVal(Unknown)})
}