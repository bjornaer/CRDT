echo "adding item01 to server 01 \n"
curl -X POST http://localhost:8080/item -H 'Content-Type: application/json' -d '{"item":"item01"}'
echo "\n"
echo "adding item02 to server 01 \n"
curl -X POST http://localhost:8080/item -H 'Content-Type: application/json' -d '{"item":"item02"}'
echo "\n"
echo "deleting item01 from server 02\n"
curl -X DELETE http://localhost:8081/item -H 'Content-Type: application/json' -d '{"item":"item01"}'
echo "adding item03 to server 03 \n"
curl -X POST http://localhost:8080/item -H 'Content-Type: application/json' -d '{"item":"item03"}'
echo "\n"
echo 'Expected list on both nodes: ["item02","item03"]'
echo "get consistent list from server 02\n"
curl http://localhost:8081/item
echo "\n"
echo "get consistent list from server 01\n"
curl http://localhost:8080/item
echo "\n"
