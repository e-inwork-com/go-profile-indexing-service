package data

type Indexes struct {
	Profiles ProfileIndex
}

func InitIndexes(solrURL string, solrProfile string) Indexes {
	return Indexes{
		Profiles: ProfileIndex{SolrURL: solrURL, SolrProfile: solrProfile},
	}
}
