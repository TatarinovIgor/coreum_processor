# Coreum processing
Coreum processing is a part of solution 
it contains of the following modules:
1. [ory/kratos](https://www.ory.sh/kratos/) - authentication and user management
2. coreum_processor - an example of crypto processing on the base of [Coreum](https://www.coreum.com/) blockchain 
3. coreum_multisign_service - an example of the services to implement multi-signature from merchant infrastructure

## How to start locally
1. To build coreum processor you should run the following docker command:
```
   docker build . -t birdhousedockers/coreum_processor:latest -f ./Dockerfile
```
2. To build coreum mulrisign service you should run the following docker command
```
   docker build . -t birdhousedockers/coreum_multisign_service:latest -f ./Dockerfile-MS
```
3. Go to folder ./ory where located files to run docker compose
- if you would like to run Coreum processing on you local computer run the following command:
```
docker compose -f docker-compose.yml up
```
- if you would like to run Coreum from IDE you should specify requried env variable 
to run coreum_processor and coreum_multisign_service and use the following docker compose command:
```
docker compose -f docker-compose-local.yml up
```
***
provided docker compose files are responsible to start and run required infrastructure components like 
postgres, ory/kratos, and run required migration scripts for database </br>
if docker-compose.yml was started successfully the coreum_processor component will respond on address http://127.0.0.1:9090
***
### Coreum processing ENV variable

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

### Coreum multi-signature service ENV variable
the following env variables should be provided to run coreum multi-signature service

| name         | example                                                                                                                                                         | description                                         |
|--------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| PORT         | 9095                                                                                                                                                            | port that used coreum processing to recive requests |
| PUBLIC_KEY   | ./cmd/cryptoProcessorKey.key.pub                                                                                                                                | path to a file with public key to verify JWT        |
| MNEMONICS    | innocent beyond seed awful program shiver link flat february claw focus glimpse canvas slush forest code rough emotion juice another satisfy boil dutch unknown | mnemonic for multisignature operation               |
| NETWORK_TYPE | Testnet                                                                                                                                                         | type of Coreum network                              |

## Coreum processing user interface

### Registration of first user as admin with default merchant

### Registration of second user and new merchant

### Approval for new merchant

### Activation of new merchant and request of FT assets

### Approval of requested FT assets (issuing requested FT)

### Transfer issued FT from receiving to transferring wallets


