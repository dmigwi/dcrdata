package internal

const (
	// Tickets

	CreateTicketsTable = `CREATE TABLE IF NOT EXISTS tickets (  
		id SERIAL PRIMARY KEY,
		tx_hash TEXT NOT NULL,
		block_hash TEXT NOT NULL,
		block_height INT4,
		purchase_tx_db_id INT8,
		stakesubmission_address TEXT,
		is_multisig BOOLEAN,
		is_split BOOLEAN,
		num_inputs INT2,
		price FLOAT8,
		fee FLOAT8,
		spend_type INT2,
		pool_status INT2,
		spend_height INT4,
		spend_tx_db_id INT8,
		block_time INT8,
		agenda_id TEXT,
		agenda_vote_choice TEXT
	);`

	// Insert
	insertTicketRow0 = `INSERT INTO tickets (
		tx_hash, block_hash, block_height, purchase_tx_db_id,
		stakesubmission_address, is_multisig, is_split,
		num_inputs, price, fee, spend_type, pool_status)
	VALUES (
		$1, $2, $3,	$4,
		$5, $6, $7,
		$8, $9, $10, $11, $12) `
	insertTicketRow = insertTicketRow0 + `RETURNING id;`
	// insertTicketRowChecked = insertTicketRow0 + `ON CONFLICT (tx_hash, block_hash) DO NOTHING RETURNING id;`
	upsertTicketRow = insertTicketRow0 + `ON CONFLICT (tx_hash, block_hash) DO UPDATE 
		SET tx_hash = $1, block_hash = $2 RETURNING id;`
	insertTicketRowReturnId = `WITH ins AS (` +
		insertTicketRow0 +
		`ON CONFLICT (tx_hash, block_hash) DO UPDATE
		SET tx_hash = NULL WHERE FALSE
		RETURNING id
		)
	SELECT id FROM ins
	UNION  ALL
	SELECT id FROM tickets
	WHERE  tx_hash = $1 AND block_hash = $2
	LIMIT  1;`

	SelectTicketsInBlock         = `SELECT * FROM tickets WHERE block_hash = $1;`
	SelectTicketsTxDbIDsInBlock  = `SELECT purchase_tx_db_id FROM tickets WHERE block_hash = $1;`
	SelectTicketsForAddress      = `SELECT * FROM tickets WHERE stakesubmission_address = $1;`
	SelectTicketsForPriceAtLeast = `SELECT * FROM tickets WHERE price >= $1;`
	SelectTicketsForPriceAtMost  = `SELECT * FROM tickets WHERE price <= $1;`
	SelectTicketIDHeightByHash   = `SELECT id, block_height FROM tickets WHERE tx_hash = $1;`
	SelectTicketIDByHash         = `SELECT id FROM tickets WHERE tx_hash = $1;`
	SelectTicketStatusByHash     = `SELECT id, spend_type, pool_status FROM tickets WHERE tx_hash = $1;`
	SelectUnspentTickets         = `SELECT id, tx_hash FROM tickets WHERE spend_type = 0 OR spend_type = -1;`

	// Update
	SetTicketSpendingInfoForHash = `UPDATE tickets
		SET spend_type = $5, spend_height = $3, spend_tx_db_id = $4, pool_status = $6
		WHERE tx_hash = $1 and block_hash = $2;`
	SetTicketSpendingInfoForTicketDbID = `UPDATE tickets
		SET spend_type = $4, spend_height = $2, spend_tx_db_id = $3, pool_status = $5
		WHERE id = $1;`
	SetTicketSpendingInfoForTxDbID = `UPDATE tickets
		SET spend_type = $4, spend_height = $2, spend_tx_db_id = $3, pool_status = $5
		WHERE purchase_tx_db_id = $1;`
	SetTicketPoolStatusForTicketDbID = `UPDATE tickets SET pool_status = $2 WHERE id = $1;`
	SetTicketPoolStatusForHash       = `UPDATE tickets SET pool_status = $2 WHERE tx_hash = $1;`

	// Index
	IndexTicketsTableOnHashes = `CREATE UNIQUE INDEX uix_ticket_hashes_index
		ON tickets(tx_hash, block_hash);`
	DeindexTicketsTableOnHashes = `DROP INDEX uix_ticket_hashes_index;`

	IndexTicketsTableOnTxDbID = `CREATE UNIQUE INDEX uix_ticket_ticket_db_id
		ON tickets(purchase_tx_db_id);`
	DeindexTicketsTableOnTxDbID = `DROP INDEX uix_ticket_ticket_db_id;`

	DeleteTicketsDuplicateRows = `DELETE FROM tickets
		WHERE id IN (SELECT id FROM (
				SELECT id, ROW_NUMBER()
				OVER (partition BY tx_hash, block_hash ORDER BY id) AS rnum
				FROM tickets) t
			WHERE t.rnum > 1);`

	// Votes

	CreateVotesTable = `CREATE TABLE IF NOT EXISTS votes (
		id SERIAL PRIMARY KEY,
		height INT4,
		tx_hash TEXT NOT NULL,
		block_hash TEXT NOT NULL,
		candidate_block_hash TEXT NOT NULL,
		version INT2,
		vote_bits INT2,
		block_valid BOOLEAN,
		ticket_hash TEXT,
		ticket_tx_db_id INT8,
		ticket_price FLOAT8,
		vote_reward FLOAT8
	);`

	// Insert
	insertVoteRow0 = `INSERT INTO votes (
		height, tx_hash,
		block_hash, candidate_block_hash,
		version, vote_bits, block_valid,
		ticket_hash, ticket_tx_db_id, ticket_price, vote_reward)
	VALUES (
		$1, $2,
		$3, $4,
		$5, $6, $7,
		$8, $9, $10, $11) `
	insertVoteRow = insertVoteRow0 + `RETURNING id;`
	// insertVoteRowChecked = insertVoteRow0 + `ON CONFLICT (tx_hash, block_hash) DO NOTHING RETURNING id;`
	upsertVoteRow = insertVoteRow0 + `ON CONFLICT (tx_hash, block_hash) DO UPDATE 
		SET tx_hash = $2, block_hash = $3 RETURNING id;`
	insertVoteRowReturnId = `WITH ins AS (` +
		insertVoteRow0 +
		`ON CONFLICT (tx_hash, block_hash) DO UPDATE
		SET tx_hash = NULL WHERE FALSE
		RETURNING id
		)
	SELECT id FROM ins
	UNION  ALL
	SELECT id FROM votes
	WHERE  tx_hash = $2 AND block_hash = $3
	LIMIT  1;`

	SelectAllVoteDbIDsHeightsTicketHashes = `SELECT id, height, ticket_hash FROM votes;`
	SelectAllVoteDbIDsHeightsTicketDbIDs  = `SELECT id, height, ticket_tx_db_id FROM votes;`

	// Index
	IndexVotesTableOnHashes = `CREATE UNIQUE INDEX uix_votes_hashes_index
		ON votes(tx_hash, block_hash);`
	DeindexVotesTableOnHashes = `DROP INDEX uix_votes_hashes_index;`

	IndexVotesTableOnCandidate = `CREATE INDEX uix_votes_candidate_block
		ON votes(candidate_block_hash);`
	DeindexVotesTableOnCandidate = `DROP INDEX uix_votes_candidate_block;`

	IndexVotesTableOnVoteVersion = `CREATE INDEX uix_votes_vote_version
		ON votes(version);`
	DeindexVotesTableOnVoteVersion = `DROP INDEX uix_votes_vote_version;`

	DeleteVotesDuplicateRows = `DELETE FROM votes
		WHERE id IN (SELECT id FROM (
				SELECT id, ROW_NUMBER()
				OVER (partition BY tx_hash, block_hash ORDER BY id) AS rnum
				FROM votes) t
			WHERE t.rnum > 1);`

	// Misses

	CreateMissesTable = `CREATE TABLE IF NOT EXISTS misses (
		id SERIAL PRIMARY KEY,
		height INT4,
		block_hash TEXT NOT NULL,
		candidate_block_hash TEXT NOT NULL,
		ticket_hash TEXT NOT NULL
	);`

	// Insert
	insertMissRow0 = `INSERT INTO misses (
		height, block_hash, candidate_block_hash, ticket_hash)
	VALUES (
		$1, $2, $3, $4) `
	insertMissRow = insertMissRow0 + `RETURNING id;`
	// insertVoteRowChecked = insertMissRow0 + `ON CONFLICT (ticket_hash, block_hash) DO NOTHING RETURNING id;`
	upsertMissRow = insertMissRow0 + `ON CONFLICT (ticket_hash, block_hash) DO UPDATE 
		SET ticket_hash = $4, block_hash = $2 RETURNING id;`
	insertMissRowReturnId = `WITH ins AS (` +
		insertMissRow0 +
		`ON CONFLICT (ticket_hash, block_hash) DO UPDATE
		SET ticket_hash = NULL WHERE FALSE
		RETURNING id
		)
	SELECT id FROM ins
	UNION  ALL
	SELECT id FROM misses
	WHERE  ticket_hash = $4 AND block_hash = $2
	LIMIT  1;`

	SelectMissesInBlock = `SELECT ticket_hash FROM misses WHERE block_hash = $1;`

	// Index
	IndexMissesTableOnHashes = `CREATE UNIQUE INDEX uix_misses_hashes_index
		ON misses(ticket_hash, block_hash);`
	DeindexMissesTableOnHashes = `DROP INDEX uix_misses_hashes_index;`

	DeleteMissesDuplicateRows = `DELETE FROM misses
		WHERE id IN (SELECT id FROM (
				SELECT id, ROW_NUMBER()
				OVER (partition BY ticket_hash, block_hash ORDER BY id) AS rnum
				FROM misses) t
			WHERE t.rnum > 1);`

	// Revokes?

	// Agendas
	CreateAgendasTable = `CREATE TABLE IF NOT EXISTS agendas (
		id SERIAL PRIMARY KEY,
		agenda_id TEXT,
		agenda_vote_choice TEXT,
		tx_hash TEXT NOT NULL,
		block_height INT4,
		block_time INT8
	);`

	// Insert
	insertAgendaRow0 = `INSERT INTO agendas (
		agenda_id, agenda_vote_choice,
		tx_hash, block_height, block_time)
		VALUES ($1, $2, $3, $4, $5) `

	insertAgendaRow = insertAgendaRow0 + `RETURNING id;`

	upsertAgendaRow = insertAgendaRow0 + `ON CONFLICT (agenda_id, agenda_vote_choice, tx_hash, block_height) DO UPDATE 
		SET block_time = $5 RETURNING id;`

	SelectAgendasAgendaVotes = `SELECT to_timestamp(block_time)::date as date,
		COUNT(CASE WHEN agenda_vote_choice = 'yes' THEN 1 ELSE NULL END) as yes,
		COUNT(CASE WHEN agenda_vote_choice = 'no' THEN 1 ELSE NULL END) as no,
		COUNT(CASE WHEN agenda_vote_choice = 'abstain' THEN 1 ELSE NULL END) as abstain,
		count(*) as total FROM agendas WHERE agenda_id = $1
		GROUP BY date ORDER BY date;`

	IndexAgendasTableOnBlockTime = `CREATE INDEX uix_agendas_block_time
		ON agendas(block_time);`
	DeindexAgendasTableOnBlockTime = `DROP INDEX uix_agendas_block_time;`

	IndexAgendasTableOnAgendaID = `CREATE UNIQUE INDEX uix_agendas_agenda_id
		ON agendas(agenda_id, agenda_vote_choice, tx_hash, block_height);`

	DeindexAgendasTableOnAgendaID = `DROP INDEX uix_agendas_agenda_id;`
)

func MakeTicketInsertStatement(checked bool) string {
	if checked {
		return upsertTicketRow
	}
	return insertTicketRow
}

func MakeVoteInsertStatement(checked bool) string {
	if checked {
		return upsertVoteRow
	}
	return insertVoteRow
}

func MakeMissInsertStatement(checked bool) string {
	if checked {
		return upsertMissRow
	}
	return insertMissRow
}

func MakeAgendaInsertStatement(checked bool) string {
	if checked {
		return upsertAgendaRow
	}
	return insertAgendaRow
}
