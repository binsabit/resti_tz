# Account Transactions Manager

This project is an Account Transactions Manager that allows you to create accounts and manage transactions including incomes, outcomes, and transfers. The application uses PostgreSQL as the database and is containerized using Docker Compose.

## Features

- Create accounts with initial balance.
- Retrieve all accounts with their current balance.
- Create transactions (income, outcome, transfer).
- Retrieve all transactions with account names and types of operations.

## Prerequisites

- Docker
- Docker Compose

## Getting Started

Follow these steps to get the application up and running.

### 1. Clone the Repository

```bash
git clone https://github.com/binsabit/resti-tz.git
cd resti-tz
```

### 2. Configure config.yaml as such

```bash
env: "dev"
http:
  port: 8000
database:
  host: "database"
  user: "postgres"
  password: "postgres"
  name: "postgres"
  port: "5432"
```

### 3. Install Docker and run command

```bash
docker compose up -d --build
```