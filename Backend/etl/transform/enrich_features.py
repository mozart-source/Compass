import pandas as pd


def add_time_features(df: pd.DataFrame, timestamp_col='timestamp'):
    if timestamp_col in df.columns:
        ts = pd.to_datetime(df[timestamp_col], errors='coerce')
        df['day_of_week'] = ts.dt.day_name()
        df['hour_of_day'] = ts.dt.hour
        df['is_weekend'] = ts.dt.dayofweek >= 5
        df['month'] = ts.dt.month
        df['year'] = ts.dt.year
    return df
