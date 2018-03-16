
# Why
A horizontally (infinitely) scalable webscraper in Go. Based on Cassandra, Traefik, colly, gRPC, and my other [boilerplate](https://github.com/dioptre/gtrpc).

# Instructions
*Read* and run build.sh for instructions on building and executing. Just edit /service/scrape.go to customize what you want to upload to Cassandra and how. Then run for example:
```
./gcli https://en.wikipedia.org/wiki/List_of_HTTP_status_codes _ ".*wikipedia\.org.*"
```



# Thanks
Cheers to the engineers of Cassandra, colly, gRPC,Consul, Traefik & protobuf to name a few.

