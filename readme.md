### Build and Run the Application
To build and run the application, open a terminal in the project directory and run the following command:
`docker-compose up --build`
This command will build the Go API and Postgres containers and start them. You should see the logs for both services in the terminal.


### Test the API
Once the containers are up and running, you can test the API using a tool like Postman or cURL. Here are some example requests:
* GET http://localhost:8000/men - Get all books
* GET http://localhost:8000/men/{rank} - Get a man by RANK
* POST http://localhost:8000/men - Create a new man
* PUT http://localhost:8000/men/{rank} - Update a man by RANK
* DELETE http://localhost:8000/men/{rank} - Delete a man by RANK