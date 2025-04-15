import os
from pymongo import MongoClient
import pandas as pd


def get_client():
    uri = os.getenv(
        "MONGODB_URI", "mongodb+srv://ahmedelhadi1777:fb5OpNipjvS65euk@cluster0.ojy4aft.mongodb.net/compass_ai?retryWrites=true&w=majority&appName=Cluster0")
    username = os.getenv("MONGODB_USERNAME")
    password = os.getenv("MONGODB_PASSWORD")
    client_kwargs = {}
    if username and password:
        client_kwargs["username"] = username
        client_kwargs["password"] = password
    return MongoClient(uri, maxPoolSize=20, minPoolSize=5, **client_kwargs)


def fetch_collection(coll):
    """Fetch all documents from a given collection as a DataFrame."""
    client = get_client()
    db = client.get_default_database()
    return pd.DataFrame(list(db[coll].find()))


def fetch_ai_models():
    """Fetch all AI models from the ai_models collection."""
    client = get_client()
    db = client.get_default_database()
    return pd.DataFrame(list(db["ai_models"].find()))


def fetch_conversations():
    """Fetch all conversations from the conversations collection."""
    client = get_client()
    db = client.get_default_database()
    return pd.DataFrame(list(db["conversations"].find()))


def fetch_cost_tracking():
    """Fetch all cost tracking entries from the cost_tracking collection."""
    client = get_client()
    db = client.get_default_database()
    return pd.DataFrame(list(db["cost_tracking"].find()))


def flatten_dict_column(df, col):
    """Flatten a column of dicts in a DataFrame, joining keys as col_key."""
    if col in df.columns:
        dict_df = df[col].apply(pd.Series)
        dict_df = dict_df.add_prefix(f"{col}_")
        df = pd.concat([df.drop(columns=[col]), dict_df], axis=1)
    return df
