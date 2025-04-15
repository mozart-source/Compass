import pandas as pd


def clean_text_fields(df: pd.DataFrame, columns):
    for col in columns:
        if col in df.columns:
            df[col] = df[col].astype(str).str.strip(
            ).str.replace(r'\s+', ' ', regex=True)
    return df


def fill_missing(df: pd.DataFrame, fill_map: dict):
    for col, val in fill_map.items():
        if col in df.columns:
            df[col] = df[col].fillna(val)
    return df


def lowercase_column(df: pd.DataFrame, col: str):
    if col in df.columns:
        df[col] = df[col].astype(str).str.lower()
    return df


def normalize_categorical(df: pd.DataFrame, columns):
    for col in columns:
        if col in df.columns:
            df[col] = df[col].astype(str).str.lower().str.replace(' ', '_')
    return df


def strip_all_string_columns(df: pd.DataFrame):
    for col in df.select_dtypes(include=['object']).columns:
        df[col] = df[col].astype(str).str.strip()
    return df


def fill_all_missing(df: pd.DataFrame):
    # Fill NaN in string columns with '' and numeric columns with 0
    for col in df.columns:
        if df[col].dtype == 'object':
            df[col] = df[col].fillna('')
        elif pd.api.types.is_numeric_dtype(df[col]):
            df[col] = df[col].fillna(0)
    return df
