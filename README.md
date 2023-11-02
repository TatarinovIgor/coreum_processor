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
docker compose -f docker-compose.yml up -d
```
- if you would like to run Coreum from IDE you should specify requried env variable 
to run coreum_processor and coreum_multisign_service and use the following docker compose command:
```
docker compose -f docker-compose-local.yml up -d
```
***
provided docker compose files are responsible to start and run required infrastructure components like 
postgres, ory/kratos, and run required migration scripts for database </br>
if docker-compose.yml was started successfully the coreum_processor component 
will respond on address http://127.0.0.1:9090 and coreum_multisign_service on address  http://127.0.0.1:9095
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
1. open landing by address http://127.0.0.1:9090</br>
   ![Screenshot of landing](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/001-landing.png)
2. Push "Register" and fill registration form </br>
   ![Screenshot of signup](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/002-signup.png)
   ![Screenshot of verify](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/003-verify.png)

3. After "Sign up" open a page with MailSlurper that deployed by docker compose at address http://127.0.0.1:4436 </br>
   ![Screenshot of smtp](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/004-smtp.png)
 
4. Return page with verification form and put code to the form</br>
   ![Screenshot of verification](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/005-verification.png)

5. Push "Continue" and fill personal information in registration wizard</br>
   ![Screenshot of personal data](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/006-personaldata.png)

6. On next registration wizard page put the following merchant id `aaef4567-b438-48a4-9a3a-f3a730b0e1ec` 
to link new registered client with default merchant created by migration scripts </br>
   ![Screenshot of default merchant](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/007-defaultmerchant.png)

7. On dashboard page push 'Create' button to activate the first merchant </br>
   ![Screenshot of dashboard](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/008-dashboard.png) </br>

8. Do logout in order continue scenarios </br>
    ![Screenshot of activation](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/009-activation.png)

9. Login to postgres container to update the register user access rights by using the following command:
```
docker exec -it ory-db-psql-1 psql -d coreum_processor -U postgres
```
- find user id database
```
select * from users;
```
- update user access in database
```
update users set access = 4099 where identity = '<put_your_user_identity>';
```
### Registration of second user and new merchant
10. Do steps 1 - 5 from [Registration of first user as admin with default merchant](https://github.com/TatarinovIgor/coreum_processor#registration-of-first-user-as-admin-with-default-merchant)

11. on step 6 push option use option "I would like to make new merchant"
    ![Screenshot of new merchant](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/010-newmerchant.png)

12. Provide information about merchant
    ![Screenshot of merchant data](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/011-merchantdata.png)

12. Merchant approval pending
    ![Screenshot of merchant pending](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/012-merchantpending.png)

13. Do LogOut by using the following link http://127.0.0.1:9090/logout

### Approval for new merchant t
14. Login as administrator and approve merchant on the page http://127.0.0.1:9090/ui/admin/merchant-requests </br>
    ![Screenshot of merchant approval](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/013-approvemerchant.png)

15. Do logout from administrator 
### Activation of new merchant and request of FT assets
16. Login under newly approved merchant and switch to settings page
    ![Screenshot of go to settings](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/014-gotomechantsettings.png)

17. In setting page put callback url for mutli-signature service (if docker-compose.yaml was user the correct address should be `http://host.docker.internal:9095')</br>
    ![Screenshot of provide url for merchant multi-sign callbeack](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/015-putmerchantcallback.png)

18. Goto dashboard page and create Coreum Wallet
    ![Screenshot of merchant dashaboard](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/016-gotomerchantdashboard.png)

19. After successful activation the multi-sign service should have a record in logs:
```

2023-11-02T14:17:06.300072200Z 2023/11/02 14:17:06.299951 /go/src/app/cmd/multisign-service/service/multisign-service.go:133: On blockchain: coreum
2023-11-02T14:17:06.300112200Z 	for external id: 3b23b632-1b9a-4882-8713-271552c0515e-R
2023-11-02T14:17:06.300115400Z 	Given the following addresses:
2023-11-02T14:17:06.300117300Z 	 testcore1jug4e9pw92zc8vhcewdtnj3v0zp6r2c0e2hc3e
2023-11-02T14:17:06.300119300Z 
2023-11-02T14:17:10.324532500Z 2023/11/02 14:17:10.324398 /go/src/app/cmd/multisign-service/handler/handler.go:60: On blockchain: coreum 
2023-11-02T14:17:10.324567500Z  for external id: 3b23b632-1b9a-4882-8713-271552c0515e-R 
2023-11-02T14:17:10.324570700Z  Sign the following transaction: activate-ms-testcore1l9fp9j8sfzv88cwuuzd62ek8ed59l742c04ge4
2023-11-02T14:17:13.173877800Z 2023/11/02 14:17:13.173762 /go/src/app/cmd/multisign-service/service/multisign-service.go:133: On blockchain: coreum
2023-11-02T14:17:13.173909400Z 	for external id: 3b23b632-1b9a-4882-8713-271552c0515e-S
2023-11-02T14:17:13.173913700Z 	Given the following addresses:
2023-11-02T14:17:13.173916500Z 	 testcore1jug4e9pw92zc8vhcewdtnj3v0zp6r2c0e2hc3e
2023-11-02T14:17:13.173919200Z 
2023-11-02T14:17:16.347297500Z 2023/11/02 14:17:16.347139 /go/src/app/cmd/multisign-service/handler/handler.go:60: On blockchain: coreum 
2023-11-02T14:17:16.347329300Z  for external id: 3b23b632-1b9a-4882-8713-271552c0515e-S 
2023-11-02T14:17:16.347333000Z  Sign the following transaction: activate-ms-testcore1fecutgj7t0vvtll4psv0965z2j30s6k4h2vgzf
2023-11-02T14:17:34.774034800Z 2023/11/02 14:17:34.773910 /go/src/app/cmd/multisign-service/service/multisign-service.go:133: On blockchain: 
2023-11-02T14:17:34.774079300Z 	for external id: 
2023-11-02T14:17:34.774083100Z 	Given the following addresses:
2023-11-02T14:17:34.774085900Z 	 testcore1jug4e9pw92zc8vhcewdtnj3v0zp6r2c0e2hc3e
2023-11-02T14:17:34.774088400Z 

```
   ![Screenshot of merchant dashaboard](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/017-merchantwalletactivation.png)

20. Goto Assets page and requests a FT, by filling required fields
    ![Screenshot of merchant dashaboard](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/018-merchantrequestft.png)

21. Logout from the merchant

### Approval of requested FT assets (issuing requested FT)
22. Goto assets page and approve requested token
    ![Screenshot of merchant dashaboard](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/016-gotomerchantdashboard.png)

23. Switch to merchant panel to generate a deposit wallet</br>
    ![Screenshot of switch merchant panel](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/023-switchtomerchant.png)

24. Create a deposit request by providing unique id for deposit user and push "Deposit"</br>
    ![Screenshot of switch merchant panel](https://github.com/TatarinovIgor/coreum_processor/blob/main/documentation/images/024-makedeposit.png)</br>
    after generation of wallet address you should store it for following steps

25. Logout from Administrator

### Transfer issued FT from receiving to transferring wallets



