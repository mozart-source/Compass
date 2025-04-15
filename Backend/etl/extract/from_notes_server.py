import os
from pymongo import MongoClient
import pandas as pd


def get_notes_server_client():
    uri = os.getenv("NOTES_MONGODB_URI",
                    "mongodb+srv://ahmedelhadi1777:fb5OpNipjvS65euk@cluster0.ojy4aft.mongodb.net/compass_notes?retryWrites=true&w=majority&appName=Cluster0")
    username = os.getenv("NOTES_MONGODB_USERNAME")
    password = os.getenv("NOTES_MONGODB_PASSWORD")
    client_kwargs = {}
    if username and password:
        client_kwargs["username"] = username
        client_kwargs["password"] = password
    return MongoClient(uri, maxPoolSize=20, minPoolSize=5, **client_kwargs)


def fetch_notes():
    """Fetch all notes from the notes collection as a DataFrame."""
    client = get_notes_server_client()
    db = client.get_default_database()
    return pd.DataFrame(list(db["notes"].find()))


def fetch_journals():
    """Fetch all journals from the journals collection as a DataFrame."""
    client = get_notes_server_client()
    db = client.get_default_database()
    return pd.DataFrame(list(db["journals"].find()))


def flatten_list_column(df, col):
    """Flatten a column of lists in a DataFrame, joining values as comma-separated string."""
    if col in df.columns:
        df[col] = df[col].apply(lambda x: ','.join(
            map(str, x)) if isinstance(x, list) else x)
    return df
