package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Khan/genqlient/graphql"
	"github.com/bread-amammay/jiragraphql/jira/graphqlgen"
)

type authedTransport struct {
	user, password string
	wrapped        http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.user, t.password)
	return t.wrapped.RoundTrip(req)
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("error running: %v", err)
	}
	log.Println("done")
}

func run() error {

	user := os.Getenv("JIRA_USER")
	if user == "" {
		return fmt.Errorf("JIRA_USER must be set")
	}

	secret := os.Getenv("JIRA_SECRET")
	if secret == "" {
		return fmt.Errorf("JIRA_SECRET must be set")
	}

	httpClient := &http.Client{
		Transport: &authedTransport{
			user:     user,
			password: secret,
			wrapped:  http.DefaultTransport,
		},
	}

	graphqlClient := graphql.NewClient("https://breadfinance.atlassian.net/gateway/api/graphql", httpClient)

	var (
		orgID  = "ari:cloud:platform::org/39bbec45-1df7-4e37-a53b-e7f379c8f3a6"
		siteID = "60507299-4045-4168-bafb-7d5e7d3356f1"
		first  = 50
		sortBy = []graphqlgen.TeamSort{{
			Field: "DISPLAY_NAME",
			Order: "ASC",
		}}
	)

	teams, err := getAllTeams(context.Background(), graphqlClient, orgID, siteID, first, sortBy)
	if err != nil {
		return err
	}

	for _, team := range teams {
		log.Printf("***team: %s\n", team.teamName)
		for _, member := range team.members {
			log.Printf("member: %s\n", member)
		}
	}
	return nil

}

type Team struct {
	teamName string
	members  []string
}

func getAllTeams(ctx context.Context, client graphql.Client, orgID, siteID string, first int, sortBy []graphqlgen.TeamSort) ([]Team, error) {

	var allTeams []Team
	var teamAfter string

	for {
		teamWithMembers, err := graphqlgen.TeamWithMembers(ctx, client, orgID, siteID, first, sortBy, "", teamAfter)
		if err != nil {
			return nil, err
		}

		for _, team := range teamWithMembers.Team.TeamSearchV2.Nodes {
			var members []string
			for _, member := range team.Team.Members.Nodes {
				members = append(members, member.Member.GetName())
			}
			allTeams = append(allTeams, Team{teamName: team.Team.DisplayName, members: members})
		}

		if !teamWithMembers.Team.TeamSearchV2.PageInfo.HasNextPage {
			break
		}

		teamAfter = teamWithMembers.Team.TeamSearchV2.PageInfo.EndCursor
	}

	return allTeams, nil
}
