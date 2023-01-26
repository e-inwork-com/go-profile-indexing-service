package data

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/golang-module/carbon"
)

type ProfileIndex struct {
	SolrURL     string
	SolrProfile string
}

func (i ProfileIndex) Update(profile *Profile) (*http.Response, error) {
	if !profile.IsDeleted {
		createAt := carbon.Parse(profile.CreatedAt.String()).ToRfc3339String("UTC")
		record := fmt.Sprintf(
			`{"id": "%v", "created_at":  "%v", "profile_user": "%v", "profile_name": "%v", "profile_picture": "%v", "version": "%v"}`,
			profile.ID, createAt, profile.ProfileUser, profile.ProfileName, profile.ProfilePicture, profile.Version)

		res, err := http.Post(
			i.SolrURL+"/api/collections/"+i.SolrProfile+"/update?commit=true",
			"application/json",
			bytes.NewReader([]byte(record)))

		if err != nil {
			return res, err
		}
		if res.StatusCode != http.StatusOK {
			return res, res.Request.Context().Err()
		}

		return res, nil
	} else {
		res, err := http.Post(
			i.SolrURL+"/solr/"+i.SolrProfile+"/update?commit=true",
			"application/json",
			bytes.NewReader([]byte("{'delete': {'query': 'id:"+profile.ID.String()+"'}}")))

		if err != nil {
			return res, err
		}
		if res.StatusCode != http.StatusOK {
			return res, res.Request.Context().Err()
		}

		return res, nil
	}
}
