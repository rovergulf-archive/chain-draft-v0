# Consensus

### Nodes

- **Network Validator** - TBD. Nodes maintained by Rovergulf Engineers.
- **Full Chain Validator** - This node type keeps all the chain blocks and transactions. Peering award is most high.
- **Address Validator** - Keeps only block headers and transactions related to this node address and its accounts. 
- **Ledger** - TBD

### Algorithm

1) Account 0x1 creates transaction to send 10 Coins (10e9 Nether) to 0x2.
2) Transaction applies a fee and returns a receipt to sender or a requester
3) Node that received a new Tx request sends it to network, to sync a new block state
4) Each 15 seconds network servers randomly prepares a node which would be applied to generate new state 
block containing all the pending transactions limited to 1024 (??)
5) Once new block generated its headers propagates via raft for each network validator peers
6) Once block header verified with its node sign, chain state updates