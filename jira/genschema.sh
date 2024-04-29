#!/bin/bash

npx get-graphql-schema@2.1.2 https://breadfinance.atlassian.net/gateway/api/graphql > schema.graphql
sed -i "" "s/input VirtualAgentPropertiesInput/input VirtualAgentPropertiesInput {\n    f: String!\n}/g" ./schema.graphql
go run github.com/Khan/genqlient@v0.7.0 genqlient.yaml
