package utils

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type KronosClient struct {
    BaseURL   string
    AuthToken string
}

func (kc *KronosClient) CreateTeamRepo(teamName string) error {
    url := fmt.Sprintf("%s/repos/team/create", kc.BaseURL)
    body := map[string]string{"team_name": teamName}
    jsonBody, _ := json.Marshal(body)

    req, _ := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
    req.Header.Set("Authorization", "Bearer "+kc.AuthToken)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("failed to create team repo: %s", resp.Status)
    }
    return nil
}

func (kc *KronosClient) AddCollaborators(repoName string, usernames []string) error {
    url := fmt.Sprintf("%s/repos/%s/collaborators/add", kc.BaseURL, repoName)
    body := map[string][]string{"usernames": usernames}
    jsonBody, _ := json.Marshal(body)

    req, _ := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
    req.Header.Set("Authorization", "Bearer "+kc.AuthToken)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("failed to add collaborators: %s", resp.Status)
    }
    return nil
}