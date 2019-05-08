function create_table_blocks () 
    box.execute("CREATE TABLE IF NOT EXISTS blocks (blockID TEXT PRIMARY KEY, height INTEGER, parentID TEXT, blockfacts TEXT)")
end

function create_table_transactions () 
    box.execute("CREATE TABLE IF NOT EXISTS transactions (txid TEXT PRIMARY KEY, blockID text, CONSTRAINT fk_blockid FOREIGN KEY(blockID) REFERENCES blocks(blockID), minerfee INTEGER, extData TEXT);")
end

function create_table_unlockconditions () 
    box.execute("CREATE TABLE IF NOT EXISTS unlockconditions (hash TEXT PRIMARY KEY, unlockcondition TEXT);")
end

function create_table_fullfillments () 
    box.execute("CREATE TABLE IF NOT EXISTS fullfillments (hash TEXT PRIMARY KEY, fullfillment TEXT);")
end

function create_table_coinoutput () 
    box.execute("CREATE TABLE IF NOT EXISTS coinoutput (id TEXT PRIMARY KEY, txid TEXT, CONSTRAINT fk_txid FOREIGN KEY(txid) REFERENCES transactions(txid), value TEXT, unlockConditionHash TEXT, FOREIGN KEY(unlockConditionHash) REFERENCES unlockconditions(hash));")
end

function create_table_coininput () 
    box.execute("CREATE TABLE IF NOT EXISTS coininput (id TEXT PRIMARY KEY, parentid TEXT, CONSTRAINT fk_parentid FOREIGN KEY(id) REFERENCES coinoutput(id), txid TEXT, CONSTRAINT fk_txid FOREIGN KEY(txid) REFERENCES transactions(txid), fullfillmentHash TEXT, FOREIGN KEY(fullfillmentHash) REFERENCES fullfillments(hash));")
end

function create_table_blockstakeoutput () 
    box.execute("CREATE TABLE IF NOT EXISTS blockstakeoutput (id TEXT PRIMARY KEY, txid TEXT, CONSTRAINT fk_txid FOREIGN KEY(txid) REFERENCES transactions(txid), value TEXT, unlockConditionHash TEXT, FOREIGN KEY(unlockConditionHash) REFERENCES unlockconditions(hash));")
end

function create_table_blockstakeinput () 
    box.execute("CREATE TABLE IF NOT EXISTS blockstakeinput (id TEXT PRIMARY KEY, parentid TEXT, CONSTRAINT fk_parentid FOREIGN KEY(id) REFERENCES coinoutput(id), txid TEXT, CONSTRAINT fk_txid FOREIGN KEY(txid) REFERENCES transactions(txid), fullfillmentHash TEXT, FOREIGN KEY(fullfillmentHash) REFERENCES fullfillments(hash));")
end

function create_table_info () 
    box.execute("CREATE TABLE IF NOT EXISTS info (consensusChangeID TEXT PRIMARY KEY, blockheight INTEGER);")
end

function createExplorerDatabase()
    create_table_blocks()
    create_table_transactions()
    create_table_unlockconditions()
    create_table_fullfillments()
    create_table_coinoutput()
    create_table_coininput()
    create_table_blockstakeoutput()
    create_table_blockstakeinput()
    create_table_info()
    return "db created"
end

