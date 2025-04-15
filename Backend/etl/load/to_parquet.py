import pandas as pd


def save_to_parquet(df: pd.DataFrame, path: str = "etl_output/combined.parquet"):
    """Save DataFrame to a Parquet file."""
    df.to_parquet(path, index=False)
