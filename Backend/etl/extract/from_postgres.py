import os
import pandas as pd
from sqlalchemy import create_engine


def get_engine():
    return create_engine(
        f"postgresql+psycopg2://{os.getenv('DB_USER','elhadi')}:{os.getenv('DB_PASSWORD','test123')}@{os.getenv('DB_HOST','localhost')}:{os.getenv('DB_PORT',5432)}/{os.getenv('DB_NAME','compass')}",
        pool_size=5,
        max_overflow=10,
        pool_timeout=30,
        pool_recycle=1800
    )


def fetch_table(table):
    engine = get_engine()
    return pd.read_sql(f"SELECT * FROM {table}", engine)


def fetch_users():
    engine = get_engine()
    return pd.read_sql("SELECT * FROM users", engine)


def fetch_projects():
    engine = get_engine()
    return pd.read_sql("SELECT * FROM projects", engine)
