# InsaneSDFS
A secure way to manage and store files on a distributed system while ensuring security and privacy.


## How it works

There are 3 components involved in this:
1. File server (stores files for all people and only is hosted locally - on port: 127.0.0.1 to ensure security)
2. File client (accesses files servers)
3. File gateway/proxy (allows file clients to access file servers without exposing the server directly and the rest of the network)

Identity on file clients will be managed via DIDs (a custom DID server which I'll provide) and OAuth2 identity providers. When creating a session for working with file server (either over mTLS via TCP or mTLS over QUIC UDP), the file client will send its ID (either via did token or JWT token), and file server will verify authentication based on users included in its database. To register a user, you must be on the same network and have access to the admin console (i.e the file server administrator must grant you access either via email link or QR code or what not), from there your user info (which includes DID verifcation method + service & OAuth2 jwks endpoint) will be registered onto the file server and you'll be able to access files in your user and in the "common"/"shared" file system. Permissions for users to manage the shared system will be managed by the file system admin (i.e maybe they'll make it read only and such that only certain users are allowed to add or remove or update files). File gateway will be used to ensure secure transmission of commands to query and mutate actions on file server will be used and performed using E2EE, ensuring that whoever manages the file gateway cannot view the content of the information, and only merely forwards traffic and assists in device discovery based on file servers registered with it.

SFTP support is too early to decide at this stage (not yet decided whether to include this). Common clients for this will include: desktop client (Linux, Windows, MacOS), web client, mobile client (both Android and Apple), CLI clients, and Library API clients.
