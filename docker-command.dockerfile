
# компиляция образа
docker build -t puzanovma/bill18test .

# запуск контейнера
docker run --rm -it -p 8091:8091 puzanovma/bill18test 
# в фоне
docker run --rm -it -d -p 8091:8091 puzanovma/bill18test 
#Остановим и удалим последний контейнер:
docker stop $(docker ps -lq) && docker rm $(docker ps -lq)

#Остановить все Docker контейнеры:
docker stop $(docker ps -a -q)
#Удалить все Docker контейнеры:
docker rm $(docker ps -a -q)

#Поднять docker-compose
docker-compose up
