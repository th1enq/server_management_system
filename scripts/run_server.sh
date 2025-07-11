echo 'Docker Setup ...'
docker-compose up --build -d
echo 'Waiting for docker ...'
sleep 10
echo 'Docker setup successfully !!!'

echo 'Migrate up database ...'
go run ./cmd/migrate up
echo 'Migrate database successfully !!!'

echo 'Server setup...'
go run ./cmd/server
echo 'Server is starting now. Try in: http://localhost:8080/swagger/index.html'