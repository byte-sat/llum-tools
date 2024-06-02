// generated @ 2024-06-02T00:06:36+03:00 by gendoc
package main

import "github.com/noonien/codoc"

func init() {
	codoc.Register(codoc.Package{
		ID:   "github.com/byte-sat/llum-tools",
		Name: "main",
		Doc:  "generated @ 2024-06-01T12:45:20+03:00 by gendoc",
		Functions: map[string]codoc.Function{
			"Whois": {
				Name: "Whois",
				Doc:  "Get domain whois\ndomain: domain name to check. e.g. example.com",
				Args: []string{
					"domain",
				},
			},
			"add": {
				Name: "add",
				Doc:  "adds two numbers\na: the first number\nb: the second number",
				Args: []string{
					"a",
					"b",
				},
			},
			"init": {
				Name: "init",
			},
			"main": {
				Name: "main",
			},
			"woop": {
				Name: "woop",
				Doc:  "woops the foo\nf: foo",
				Args: []string{
					"f",
				},
			},
		},
		Structs: map[string]codoc.Struct{
			"Foo": {
				Name: "Foo",
				Fields: map[string]codoc.Field{
					"A": {
						Comment: "foo",
					},
				},
			},
			"ToolRepo": {
				Name: "ToolRepo",
				Methods: map[string]codoc.Function{
					"GetToolSchema": {
						Name: "GetToolSchema",
						Args: []string{
							"w",
							"r",
						},
					},
					"InvokeTool": {
						Name: "InvokeTool",
						Args: []string{
							"w",
							"r",
						},
					},
				},
			},
		},
	})
}
