./bin/kafka-topics.sh --create --topic order-received  --bootstrap-server localhost:9092
./bin/kafka-configs.sh --alter --topic order-received --add-config delete.retention.ms=259200000  --bootstrap-server localhost:9092

./bin/kafka-topics.sh --create --topic order-confirmed --bootstrap-server localhost:9092
./bin/kafka-topics.sh --create --topic order-packed --bootstrap-server localhost:9092
./bin/kafka-topics.sh --create --topic email-notification --bootstrap-server localhost:9092
./bin/kafka-topics.sh --create --topic email-notification --bootstrap-server localhost:9092
./bin/kafka-topics.sh --create --topic error --bootstrap-server localhost:9092


