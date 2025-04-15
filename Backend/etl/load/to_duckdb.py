import duckdb
import pandas as pd


def save(df: pd.DataFrame, table: str = "user_events", db_path: str = "etl_output/analytics.duckdb"):
    con = duckdb.connect(db_path)
    # Create table if not exists, then insert
    con.execute(
        f"CREATE TABLE IF NOT EXISTS {table} AS SELECT * FROM df LIMIT 0")
    con.execute(f"INSERT INTO {table} SELECT * FROM df")
    con.close()
