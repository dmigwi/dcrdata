package agendadb

import (
	"fmt"
	"os"

	"github.com/decred/dcrd/rpcclient"

	"github.com/asdine/storm"
	"github.com/decred/dcrd/dcrjson"
)

type AgendaDB struct {
	sdb        *storm.DB
	NumAgendas int
	NumChoices int
}

// Open will either open and existing database or create a new one using with
// the specified file name.
var dbName = "Agendas.db"

func Open(dbPath string) (*AgendaDB, error) {
	_, err := os.Stat(dbPath)
	isNewDB := os.IsNotExist(err)

	db, err := storm.Open(dbPath)
	if err != nil {
		return nil, err
	}

	var numAgendas, numChoices int
	if !isNewDB {
		numAgendas, err = db.Count(&AgendaTagged{})
		if err != nil {
			fmt.Printf("Unable to count agendas in existing DB: %v\n", err)
		}
		numChoices, err = db.Count(&ChoiceLabeled{})
		if err != nil {
			fmt.Printf("Unable to count choices in existing DB: %v\n", err)
		}
		fmt.Printf("Opened existing datatbase with %d agendas.\n", numAgendas)
	}

	agendaDB := &AgendaDB{
		sdb:        db,
		NumAgendas: numAgendas,
		NumChoices: numChoices,
	}
	return agendaDB, err
}

// Close should be called when you are done with the AgendaDB to close the
// underlying database.
func (db *AgendaDB) Close() error {
	return db.sdb.Close()
}

// StoreAgenda saves an agenda in the database.
func (db *AgendaDB) StoreAgenda(agenda *AgendaTagged) error {
	if db == nil || db.sdb == nil {
		return fmt.Errorf("AgendaDB not initialized")
	}
	return db.sdb.Save(agenda)
}

// LoadAgenda retrieves an agenda corresponding to the specified unique agenda
// ID, or nil if it does not exist.
func (db *AgendaDB) LoadAgenda(agendaID string) (*AgendaTagged, error) {
	if db == nil || db.sdb == nil {
		return nil, fmt.Errorf("AgendaDB not initialized")
	}
	agenda := new(AgendaTagged)
	if err := db.sdb.One("Id", agendaID, agenda); err != nil {
		return nil, err
	}
	return agenda, nil
}

// ListAgendas lists all agendas stored in the database in order of StartTime.
func (db *AgendaDB) ListAgendas() error {
	if db == nil || db.sdb == nil {
		return fmt.Errorf("AgendaDB not initialized")
	}
	q := db.sdb.Select().OrderBy("StartTime")
	i := 0
	return q.Each(new(AgendaTagged), func(record interface{}) error {
		a := record.(*AgendaTagged)
		fmt.Printf("%d: %s\n", i, a.Id)
		i++
		return nil
	})
}

// AgendaTagged has the same fields as dcrjson.Agenda, but with the Id field
// marked as the primary key via the `storm:"id"` tag. Fields tagged for
// indexing by the DB are: StartTime, ExpireTime, Status, and QuorumProgress.
type AgendaTagged struct {
	Id             string           `json:"id" storm:"id"`
	Description    string           `json:"description"`
	Mask           uint16           `json:"mask"`
	StartTime      uint64           `json:"starttime" storm:"index"`
	ExpireTime     uint64           `json:"expiretime" storm:"index"`
	Status         string           `json:"status" storm:"index"`
	QuorumProgress float64          `json:"quorumprogress" storm:"index"`
	Choices        []dcrjson.Choice `json:"choices"`
	VoteVersion    uint32           `json:"voteversion"`
}

// ChoiceLabeled embeds dcrjson.Choice along with the AgendaID for the choice,
// and a string array suitable for use as a primary key. The AgendaID is tagged
// as an index for quick lookups based on the agenda.
type ChoiceLabeled struct {
	AgendaChoice   [2]string `storm:"id"`
	AgendaID       string    `json:"agendaid" storm:"index"`
	dcrjson.Choice `storm:"inline"`
}

// GetVoteAgendasForVersion is used in getting the agendas using the vote versions
func GetVoteAgendasForVersion(ver int64, client *rpcclient.Client) (agendas []AgendaTagged) {
	voteInfo, err := client.GetVoteInfo(uint32(ver))
	if err == nil {
		for _, v := range voteInfo.Agendas {
			a := AgendaTagged{
				Id:             v.Id,
				Description:    v.Description,
				Mask:           v.Mask,
				StartTime:      v.StartTime,
				ExpireTime:     v.ExpireTime,
				Status:         v.Status,
				QuorumProgress: v.QuorumProgress,
				Choices:        v.Choices,
				VoteVersion:    voteInfo.VoteVersion,
			}
			agendas = append(agendas, a)
		}
	}
	return
}

func (db *AgendaDB) CheckAvailabiltyOfVersionAgendas(version int64) bool {
	if db == nil || db.sdb == nil {
		return false
	}
	agenda := new([]AgendaTagged)
	if err := db.sdb.Find("VoteVersion", uint32(version), agenda); err != nil {
		return false
	}
	if agenda == nil {
		return false
	}
	return true
}

func (db *AgendaDB) Updatedb(voteVersion int64, client *rpcclient.Client) {
	var agendas []AgendaTagged
	for GetVoteAgendasForVersion(voteVersion, client) != nil {
		agendas = append(agendas, GetVoteAgendasForVersion(voteVersion, client)...)
		voteVersion++
	}
	for _, a := range agendas {
		err := db.StoreAgenda(&a)
		if err == nil {
			fmt.Printf("Agenda saved: %v \n", a.Id)
		}
	}
}

func CheckForUpdates(client *rpcclient.Client) {
	adb, err := Open(dbName)
	if err != nil {
		log.Errorf("Failed to open new DB: %v", err)
	}
	var voteVersion int64 = 4
	for adb.CheckAvailabiltyOfVersionAgendas(voteVersion) {
		fmt.Printf("Vote agendas exist in db for version: %v \n", voteVersion)
		voteVersion++
	}
	adb.Updatedb(voteVersion, client)
	adb.Close()
}

func GetAgendaInfo(agendaId string) (*AgendaTagged, error) {
	adb, err := Open(dbName)
	if err != nil {
		log.Errorf("Failed to open new DB: %v", err)
	}
	agenda, err := adb.LoadAgenda(agendaId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	adb.Close()
	return agenda, nil
}
