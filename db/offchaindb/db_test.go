package offchaindb

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/asdine/storm"
)

var db *storm.DB
var tempDir string

// initial sample proposal made.
var firstProposal = &ProposalInfo{
	Name:      "Initial Test proposal",
	State:     2,
	Status:    4,
	Timestamp: 1541904469,
	UserID:    "18b24b6c-14a8-45f6-ab2e-a34127840fb3",
	Username:  "secret-coder",
	PublicKey: "c7580e9d13a21a2046557f7ef0148a5be89fbe8db8c",
	Signature: "8a1b69eb08b413b3ad3161c9b43b6a65a25c537f6151866d391a352",
	Version:   "6",
	Censorship: CensorshipRecord{
		Token:      "0aaab331075d08cb03333d5a1bef04b99a708dcbfebc8f8c94040ceb1676e684",
		MerkleRoot: "cfaf772010b439db2fa175b407f7c61fc7b06fbd844192a89551abe40791b6bb",
		Signature:  "6f8a7740c518972c4dc607e877afc794be9f99a1c4790837a7104b7eb6228d4db219",
	},
	NumComments:   23,
	PublishedDate: 1541904469,
	AbandonedDate: 1543946266,
}

// TestMain sets up the temporary db needed for testing
func TestMain(m *testing.M) {
	var err error
	tempDir, err = ioutil.TempDir("", "offchain")
	if err != nil {
		log.Error(err)
		return
	}

	db, err = storm.Open(filepath.Join(tempDir, "test.db"))
	if err != nil {
		log.Error(err)
		return
	}

	//  Save the first sample proposal
	err = db.Save(firstProposal)
	if err != nil {
		log.Error(err)
		return
	}

	m.Run()

	defer db.Close()
	defer os.RemoveAll(tempDir) // clean up
}

// TestNewProposalsDB tests creating a new storm db and a http client instance.
func TestNewProposalsDB(t *testing.T) {
	var count int
	_API_URL := "https://proposals.decred.org/api/v1"
	getDbPath := func() string {
		count++
		return filepath.Join(tempDir, fmt.Sprintf("test%v.db", count))
	}

	type testData struct {
		politeiaAPIURL string
		dbPath         string

		// Checks if the db was created and its instance referenced returned.
		IsdbInstance bool
		errMsg       string
	}

	td := []testData{
		{
			politeiaAPIURL: "",
			dbPath:         "",
			IsdbInstance:   false,
			errMsg:         "missing politeia API URL",
		},
		{
			politeiaAPIURL: _API_URL,
			dbPath:         "",
			IsdbInstance:   false,
			errMsg:         "missing db path",
		},
		{
			politeiaAPIURL: "",
			dbPath:         getDbPath(),
			IsdbInstance:   false,
			errMsg:         "missing politeia API URL",
		},
		{
			politeiaAPIURL: _API_URL,
			dbPath:         getDbPath(),
			IsdbInstance:   true,
			errMsg:         "",
		},
	}

	for i, data := range td {
		t.Run("Test_#"+strconv.Itoa(i), func(t *testing.T) {
			result, err := NewProposalsDB(data.politeiaAPIURL, data.dbPath)

			var expectedErrMsg string
			if err != nil {
				expectedErrMsg = err.Error()
			}

			if expectedErrMsg != data.errMsg {
				t.Fatalf("expected to find error '%v' but found '%v'", data.errMsg, err)
			}

			if data.IsdbInstance && result != nil {
				if result._APIURLpath != _API_URL {
					t.Fatalf("expected the API URL to '%v' but found '%v'", result._APIURLpath, _API_URL)
				}

				if result.client == nil {
					t.Fatal("expected the http client not to be nil but was nil")
				}

				if result.dbP == nil {
					t.Fatal("expected the db instance not to be nil but was nil")
				}
			} else {
				// the result should be nil since the incorrect inputs resulted
				// to an error being returned and a nil proposalDB instance.
				if result != nil {
					t.Fatalf("expect the returned result to be nil but was not nil")
				}
			}
		})
	}
}

// mockServer mocks helps avoid making actual http calls during tests. It payloads
// in the same format as would be returned by the normal API endpoint.
func mockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp string
		switch r.URL.Path {
		case vettedProposalsRoute:
			resp = `{  
			"proposals":[  
			   {  
				  "name":"Change language: PoS Mining to PoS Voting, Stakepool to Voting Service Provider",
				  "state":2,
				  "status":4,
				  "timestamp":1539880429,
				  "userid":"350a4b6c-5cdd-4d87-822a-4900dc3a930c",
				  "username":"richard-red",
				  "publickey":"cd6e57b93f95dd0386d670c7ce42cb0ccd1cd5b997e87a716e9359e20251994e",
				  "signature":"c0e3285d447fd2acf1f2e1a0c86a71383dfe71b1b01e0068e56e8e7649dadb7aa503a5f99765fc3a24da8716fd5b89f75bb97762e756f15303e96d135a2e7109",
				  "files":[  
		 
				  ],
				  "numcomments":19,
				  "version":"1",
				  "publishedat":1539898457,
				  "censorshiprecord":{  
					 "token":"522652954ea7998f3fca95b9c4ca8907820eb785877dcf7fba92307131818c75",
					 "merkle":"20c9234c50e0dc78d28003fd57995192a16ca73349f5d97be456128984e463fc",
					 "signature":"d1d44788cdf8d838aad97aa829b2f27f8a32897010d6373c9d3ca1a42820dcafe2615c1904558c6628c5f9165691ead087c0cb2ada023b9aa3f76b6c587ac90e"
				  }
			   }
			]
		 }`
		case voteStatusesRoute:
			resp = `{  
		   "votesstatus":[  
			  {  
				 "token":"522652954ea7998f3fca95b9c4ca8907820eb785877dcf7fba92307131818c75",
				 "status":4,
				 "totalvotes":12745,
				 "optionsresult":[  
					{  
					   "option":{  
						  "id":"no",
						  "description":"Don't approve proposal",
						  "bits":1
					   },
					   "votesreceived":754
					},
					{  
					   "option":{  
						  "id":"yes",
						  "description":"Approve proposal",
						  "bits":2
					   },
					   "votesreceived":11991
					}
				 ],
				 "endheight":"289500",
				 "numofeligiblevotes":40958,
				 "quorumpercentage":20,
				 "passpercentage":60
			  }
		   ]
		}`
		}
		w.Write([]byte(resp))
	}))
}

// mockedPayload defines the complete unmarshalled sing payload returned by the
// mocked handleGetRequests.
var mockedPayload = &ProposalInfo{
	ID:            2,
	Name:          "Change language: PoS Mining to PoS Voting, Stakepool to Voting Service Provider",
	State:         2,
	Status:        4,
	Timestamp:     1539880429,
	UserID:        "350a4b6c-5cdd-4d87-822a-4900dc3a930c",
	Username:      "richard-red",
	PublicKey:     "cd6e57b93f95dd0386d670c7ce42cb0ccd1cd5b997e87a716e9359e20251994e",
	Signature:     "c0e3285d447fd2acf1f2e1a0c86a71383dfe71b1b01e0068e56e8e7649dadb7aa503a5f99765fc3a24da8716fd5b89f75bb97762e756f15303e96d135a2e7109",
	NumComments:   19,
	Files:         []AttachmentFile{},
	Version:       "1",
	PublishedDate: 1539898457,
	Censorship: CensorshipRecord{
		Token:      "522652954ea7998f3fca95b9c4ca8907820eb785877dcf7fba92307131818c75",
		MerkleRoot: "20c9234c50e0dc78d28003fd57995192a16ca73349f5d97be456128984e463fc",
		Signature:  "d1d44788cdf8d838aad97aa829b2f27f8a32897010d6373c9d3ca1a42820dcafe2615c1904558c6628c5f9165691ead087c0cb2ada023b9aa3f76b6c587ac90e",
	},
	VotesStatus: &ProposalVotes{
		Token:      "522652954ea7998f3fca95b9c4ca8907820eb785877dcf7fba92307131818c75",
		Status:     4,
		TotalVotes: 12745,
		VoteResults: []Results{
			{
				Option: VoteResults{
					OptionID:    "no",
					Description: "Don't approve proposal",
					Bits:        1,
				},
				VotesReceived: 754,
			},
			{
				Option: VoteResults{
					OptionID:    "yes",
					Description: "Approve proposal",
					Bits:        2,
				},
				VotesReceived: 11991,
			},
		},
		Endheight:          "289500",
		NumOfEligibleVotes: 40958,
		QuorumPercentage:   20,
		PassPercentage:     60,
	},
}

// TestStuff tests the update functionality, all proposals retrieval and proposal
// Retreival by ID.
func TestStuff(t *testing.T) {
	server := mockServer()
	newDBInstance := &ProposalDB{
		dbP:         db,
		client:      server.Client(),
		_APIURLpath: server.URL,
	}

	defer server.Close()

	// compareProposal returns an error the proposal argument passed doesn't match
	// any of the previous created proposals.
	compareProposal := func(data *ProposalInfo) error {
		if data == nil {
			return fmt.Errorf("expected the proposal not to be nil but was nil")
		}
		// data.Censorship.Token compare the proposal token against know tokens
		switch data.Censorship.Token {
		case "0aaab331075d08cb03333d5a1bef04b99a708dcbfebc8f8c94040ceb1676e684":
			if reflect.DeepEqual(data, firstProposal) {
				return nil
			}

			return fmt.Errorf("expected the initialProposal to match the retrieved but it did not")

		case "522652954ea7998f3fca95b9c4ca8907820eb785877dcf7fba92307131818c75":
			if reflect.DeepEqual(data, mockedPayload) {
				return nil
			}
			return fmt.Errorf("expected the Second Proposal to match the retrieved but it did not")

		default:
			return fmt.Errorf("unknown incorrect token found: %v", data.Censorship.Token)
		}
	}

	// Testing the update functionality
	t.Run("Test_CheckOffChainUpdates", func(t *testing.T) {
		err := newDBInstance.CheckOffChainUpdates()
		if err != nil {
			t.Fatalf("expected no error to be returned but found '%v'", err)
		}

		if newDBInstance.NumProposals != 2 {
			t.Fatalf("expected the proposals count to be 2 but found %v", newDBInstance.NumProposals)
		}
	})

	// Testing the retrieval of all proposals
	t.Run("Test_AllProposals", func(t *testing.T) {
		proposals, err := newDBInstance.AllProposals()
		if err != nil {
			t.Fatal(err)
		}

		for _, p := range proposals {
			if err = compareProposal(p); err != nil {
				t.Fatalf("expected no error but found '%v'", err)
			}
		}
	})

	// Testing proposal retrieval by ID
	t.Run("Test_ProposalByID", func(t *testing.T) {
		proposal, err := newDBInstance.ProposalByID(1)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(proposal, firstProposal) {
			t.Fatal("expected the initialProposal to match the retrieved but it did not")
		}
	})
}