# [e-inwork.com](https://e-inwork.com)

## Golang Profile Indexing Microservice
This microservice is responsible for indexing the profile data in the [Solr](https://solr.apache.org) collection.

To run this microservices, follow the command below:
1. Install Docker
    - https://docs.docker.com/get-docker/
2. Git clone this repository to your localhost, and from the terminal run below command:
   ```
   git clone git@github.com:e-inwork-com/go-profile-indexing-service
   ```
3. Change the active folder to `go-profile-indexing-service`:
   ```
   cd go-profile-service
   ```
4. Run Docker Compose:
   ```
   docker-compose -f docker-compose.local.yml up -d
   ```
5. Wait until the status of `curl-local` and `migrate-local` is `exited (0)` with the below command:
   ```
   docker-compose -f docker-compose.local.yml ps
   ```
6. Create a user in the User API with CURL command line:
    ```
    curl -d '{"email":"jon@doe.com", "password":"pa55word", "first_name": "Jon", "last_name": "Doe"}' -H "Content-Type: application/json" -X POST http://localhost:8000/service/users
    ```
7. Login to the User API:
   ```
   curl -d '{"email":"jon@doe.com", "password":"pa55word"}' -H "Content-Type: application/json" -X POST http://localhost:8000/service/users/authentication
   ```
8. You will get a token from the response login, and set it as a token variable for example like below:
   ```
   token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjhhY2NkNTUzLWIwZTgtNDYxNC1iOTY0LTA5MTYyODhkMmExOCIsImV4cCI6MTY3MjUyMTQ1M30.S-G5gGvetOrdQTLOw46SmEv-odQZ5cqqA1KtQm0XaL4
   ```
9. Create a profile for current user, you can use any image or use the image on the folder test:
   ```
   curl -F profile_name="Jon Doe" -F profile_picture=@grpc/test/images/profile.jpg -H "Authorization: Bearer $token"  -X POST http://localhost:8000/service/profiles
   ```
10. Check the Profile collection in the Solr on this link: http://localhost:8983/solr/#/profiles/query?q=*:*
11. Good luck!