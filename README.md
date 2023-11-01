# Coreum processing
Coreum processing is a part of 

## How to start locally

### Coreum processing

the following env variables should be provided to run coreum processing

| name                      | example                                                                                                                                                      | description                                          |
|---------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------|
| PORT                      | 9090                                                                                                                                                         | port that used coreum processing to recive requests  |
| TOKEN_TIME_TO_LIVE        | 300                                                                                                                                                          | time to live in seconds for JWT token                |
| PRIVATE_KEY               | ./cmd/cryptoProcessorKey                                                                                                                                     | path to a file with private key to generate JWT      |
| PUBLIC_KEY                | ./cmd/cryptoProcessorKey.key.pub                                                                                                                             | path to a file with public key to verify JWT         |
| KRATOS_URL                | http://127.0.0.1:4433                                                                                                                                        | url where kratos is hosting for user authentications |
| LISTEN_AND_SERVE_INTERVAL | 5                                                                                                                                                            | interval to listen and serve deposits                |
| DATABASE_HOST             | localhost                                                                                                                                                    | postgres host address                                |
| DATABASE_PORT             | 5438                                                                                                                                                         | postgres port                                        |
| DATABASE_NAME             | coreum_processor                                                                                                                                             | database name                                        |
| DATABASE_USER             | postgres                                                                                                                                                     | database user name                                   |
| DATABASE_PASS             | local-postgres0!                                                                                                                                             | database password                                    |
| WALLET_RECEIVER_ADDRESS   | testcore13f97kxrrq82982rsy2paqf9tx8e2jw5g2ufdfu                                                                                                              | receiver wallet address of the processing            |
| WALLET_RECEIVER_SEED      | then donate similar only tiny voyage tribe derive spare snap wet chase divide buzz play avoid captain wonder chair announce embody primary weapon breeze     | mnemonic for receiver wallet                         |
| WALLET_SENDER_ADDRESS     | testcore1w2x4hwhasqfvg8cm6kyduzgwngvp0wf46eshmc                                                                                                              | sending wallet address of the processing             |
| WALLET_SENDER_SEED        | tube pledge side laundry volume actress route pink ring galaxy vendor obscure detect patient early memory reflect glue salon valid summer scatter damp total | mnemonic for sending wallet                          |

### Coreum multi-signature service
the following env variables should be provided to run coreum multi-signature service

| name         | example                                                                                                                                                         | description                                         |
|--------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| PORT         | 9095                                                                                                                                                            | port that used coreum processing to recive requests |
| PUBLIC_KEY   | ./cmd/cryptoProcessorKey.key.pub                                                                                                                                | path to a file with public key to verify JWT        |
| MNEMONICS    | innocent beyond seed awful program shiver link flat february claw focus glimpse canvas slush forest code rough emotion juice another satisfy boil dutch unknown | mnemonic for multisignature operation               |
| NETWORK_TYPE | Testnet                                                                                                                                                         | type of Coreum network                              |
