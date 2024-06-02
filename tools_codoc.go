// generated @ 2024-06-02T21:01:42+03:00 by gendoc
package main

import "github.com/noonien/codoc"

func init() {
	codoc.Register(codoc.Package{
		ID:   "github.com/byte-sat/llum-tools",
		Name: "main",
		Doc:  "generated @ 2024-06-02T19:31:15+03:00 by gendoc",
		Functions: map[string]codoc.Function{
			"GetCID": {
				Name: "GetCID",
				Doc:  "Get the chat id\nf: foo",
				Args: []string{
					"cid",
				},
			},
			"Whois": {
				Name: "Whois",
				Doc:  "Get domain whois\ndomain: domain name to check. e.g. example.com",
				Args: []string{
					"domain",
				},
			},
			"init": {
				Name: "init",
			},
			"main": {
				Name: "main",
			},
		},
		Structs: map[string]codoc.Struct{
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
