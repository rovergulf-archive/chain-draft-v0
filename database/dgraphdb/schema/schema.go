package schema

const DefaultSchema = `
block.hash: [uid] @reverse .
block.tx_count: [uid] @count .
tx.hash: [uid] @reverse .
account.address: [uid] @reverse .
chain.length: [uid] @count .

# Define Types

type Account {
    account.address
    balance: int
    nonce: int
}

type Genesis {
    coinbase: address.address
}

type Block {
    hash: string
	genesis: string
    created_at: int
	
    starring
}

type Transaction {
    block.hash
	tx.hash
}

type Receipt {
	tx.hash
}
`
