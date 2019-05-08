--[[ Setters for inserting data in  a specific row in our database in a specific table ]]
function insert_block (blockid, height, parentid, blockfacts)
    values = {blockid, height, parentid, blockfacts}
    return box.execute([[INSERT INTO blocks VALUES (?, ?, ?, ?);]], values);
end

function insert_transaction (blockid, height, parentid, blockfacts)
    values = {blockid, height, parentid, blockfacts}
    return box.execute([[INSERT INTO transactions VALUES (?, ?, ?, ?);]], values);
end

function insert_coininput (id, parentid, txid, fullfillmentHash)
    values = {id, parentid, txid, fullfillmentHash}
    return box.execute([[INSERT INTO coininput VALUES (?, ?, ?, ?);]], values);
end

function insert_coinoutput (id, txid, value, unlockconditionHash)
    values = {id, txid, value, unlockconditionHash}
    return box.execute([[INSERT INTO coinoutput VALUES (?, ?, ?, ?);]], values);
end

function insert_blockstakeinput (id, parentid, txid, fullfillmentHash)
    values = {id, parentid, txid, fullfillmentHash}
    return box.execute([[INSERT INTO blockstakeinput VALUES (?, ?, ?, ?);]], values);
end

function insert_blockstakeoutput (id, txid, value, unlockconditionHash)
    values = {id, txid, value, unlockconditionHash}
    return box.execute([[INSERT INTO blockstakeoutput VALUES (?, ?, ?, ?);]], values);
end

function insert_fullfillment (hash, fullfillment)
    values = {hash, fullfillment}
    return box.execute([[INSERT INTO fullfillments VALUES (?, ?);]], values);
end

function insert_unlockcondition (hash, unlockcondition)
    values = {hash, unlockcondition}
    return box.execute([[INSERT INTO unlockconditions VALUES (?, ?);]], values);
end

function insert_info (consensuschangeid, height)
    values = {consensuschangeid, height}
    return box.execute([[INSERT INTO info VALUES (?, ?);]], values);
end

--[[ Getters for fetching a specific rows in our database from a specific table ]]
function get_block (blockid)
    values = {blockid}
    return box.execute([[SELECT * FROM blocks WHERE blockid = ?;]], values);
end

function get_transaction (txid)
    values = {txid}
    return box.execute([[SELECT * FROM transactions WHERE txid = ?;]], values);
end

function get_coininput (id)
    values = {id}
    return box.execute([[SELECT * FROM coininput WHERE id = ?;]], values);
end

function get_coinoutput (id)
    values = {id}
    return box.execute([[SELECT * FROM coinoutput WHERE id = ?;]], values);
end

function get_blockstakeinput (id)
    values = {id}
    return box.execute([[SELECT * FROM blockstakeinput WHERE id = ?;]], values);
end

function get_blockstakeoutput (id)
    values = {id}
    return box.execute([[SELECT * FROM blockstakeoutput WHERE id = ?;]], values);
end

function get_fullfillment (hash)
    values = {hash}
    return box.execute([[SELECT * FROM fullfillments WHERE hash = ?;]], values);
end

function get_unlockcondition (hash)
    values = {hash}
    return box.execute([[SELECT * FROM unlockconditions WHERE hash = ?;]], values);
end

function get_info (consensuschangeid)
    values = {consensuschangeid}
    return box.execute([[SELECT * FROM info WHERE consensuschangeid = ?;]], values);
end

function get_consensus_changeid ()
    return box.execute([[SELECT * FROM info]]);
end

function get_blockheight ()
    return box.execute([[SELECT TOP 1 * FROM blocks ORDER BY height DESC;]]);
end

function get_starting_blockheight ()
    return box.execute([[SELECT TOP 1 * FROM blocks ORDER BY height ASC;]]);
end

function get_transaction_block (transactionid) 
    values = {txid}
    return box.execute([[SELECT height FROM blocks JOIN transactions ON blocks.blockid = transactions.blockid WHERE transactions.txid = ?;]], values);
end

function get_unlockhash_txid_set (unlockhash) 
    values = {unlockhash}
    return box.execute([[
SELECT coininput.txid, bsinput.txid 
FROM coininput 
JOIN coinoutput ON coininput.txid = coinoutput.txid 
JOIN bsoutput on bsinput.txid = bsoutput.txid
JOIN unlockconditions on bsoutput.unlockconditionhash = unlockconditions.unlockcondition
WHERE unlockconditions.hash = ?;
]], values);
end

function get_coinoutput_txid_set (txid) 
    values = {txid}
    return box.execute([[
SELECT txid FROM transactions 
JOIN transactions ON coinoutput.txid = transactions.txid 
WHERE coinoutput.txid = ?;
]], values);
end

function get_bsoutput_txid_set (txid) 
    values = {txid}
    return box.execute([[
SELECT txid FROM transactions 
JOIN transactions ON bsoutput.txid = transactions.txid 
WHERE bsoutput.txid = ?;
]], values);
end