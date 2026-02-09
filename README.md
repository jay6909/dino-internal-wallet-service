
### Setup & Seeding
Setup & Seeding

The project includes a _/scripts/setup.sh script that initializes the database with
required seed data:

Asset types (gold, diamond, loyalty_points)

System (treasury) wallets for each asset: Treasury will have default balance of 1000000

Two users, each with one wallet per asset and an initial balance of 1000

To set up the project:

```./_scripts/setup.sh
docker-compose up -d api```


The seed process safely re-run without impacting issues with duplicacy