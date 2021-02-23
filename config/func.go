package config

import (
	"errors"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
)

func (z *ZpsConfig) configDefault() function.Function {
	return function.New(
		&function.Spec{
			Params: []function.Parameter{
				{
					Name: "value",
					Type: cty.DynamicPseudoType,
				},
				{
					Name: "default",
					Type: cty.String,
				},
			},
			Type: function.StaticReturnType(cty.String),
			Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
				if args[0].IsNull() {
					return args[1], nil
				}

				return args[0], nil
			},
		},
	)
}

func (z *ZpsConfig) coalesce() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{},
		VarParam: &function.Parameter{
			Name:             "vals",
			Type:             cty.DynamicPseudoType,
			AllowUnknown:     true,
			AllowDynamicType: true,
			AllowNull:        true,
		},
		Type: func(args []cty.Value) (ret cty.Type, err error) {
			argTypes := make([]cty.Type, len(args))
			for i, val := range args {
				argTypes[i] = val.Type()
			}
			retType, _ := convert.UnifyUnsafe(argTypes)
			if retType == cty.NilType {
				return cty.NilType, errors.New("all arguments must have the same type")
			}
			return retType, nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			for _, argVal := range args {
				// We already know this will succeed because of the checks in our
				// Type func above
				argVal, _ = convert.Convert(argVal, retType)
				if !argVal.IsKnown() {
					return cty.UnknownVal(retType), nil
				}
				if argVal.IsNull() {
					continue
				}
				if retType == cty.String && argVal.RawEquals(cty.StringVal("")) {
					continue
				}

				return argVal, nil
			}
			return cty.NilVal, errors.New("no non-null, non-empty-string arguments")
		},
	})
}
