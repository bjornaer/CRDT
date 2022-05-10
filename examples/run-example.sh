echo "adding item01 to server 01"
curl -X POST http://localhost:8080/item
   -H 'Content-Type: application/json'
   -d '{"item":"item01"}'

echo "adding item02 to server 01"
curl -X POST http://localhost:8080/item
   -H 'Content-Type: application/json'
   -d '{"item":"item02"}'

echo "deleting item01 to server 02"
curl -X DELETE http://localhost:8081/item
   -H 'Content-Type: application/json'
   -d '{"item":"item01"}'

echo "get consistent list from server 02"
curl http://localhost:8081/item
echo "get consistent list from server 01"
curl http://localhost:8080/item