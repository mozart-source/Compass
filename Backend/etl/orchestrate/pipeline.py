from etl.extract.from_notes_server import fetch_notes, fetch_journals
from etl.extract.from_postgres import fetch_table, fetch_users, fetch_projects
from etl.extract.from_mongo import fetch_collection, fetch_ai_models, fetch_conversations, fetch_cost_tracking
from etl.transform.normalize_and_join import normalize_and_join
from etl.transform.clean_fields import clean_text_fields, fill_missing, strip_all_string_columns, lowercase_column, normalize_categorical, fill_all_missing
from etl.transform.enrich_features import add_time_features
from etl.load.to_duckdb import save as save_to_duckdb
from etl.load.to_parquet import save_to_parquet
import sys
import os
import uuid
try:
    from bson import ObjectId as BsonObjectId
except ImportError:
    BsonObjectId = None
sys.path.append(os.path.dirname(os.path.dirname(__file__)))


def convert_special_types_to_str(df):
    for col in df.columns:
        if df[col].dtype == 'object':
            # Convert UUIDs
            if df[col].apply(lambda x: isinstance(x, uuid.UUID)).any():
                df[col] = df[col].apply(lambda x: str(
                    x) if isinstance(x, uuid.UUID) else x)
            # Convert ObjectIds (only if BsonObjectId is available)
            if BsonObjectId is not None:
                bson_cls = BsonObjectId
                def is_objectid(val): return val is not None and isinstance(
                    val, bson_cls)
                if df[col].apply(is_objectid).any():
                    df[col] = df[col].apply(
                        lambda x: str(x) if is_objectid(x) else x)
    return df


def run_pipeline():
    print("[ETL] Extracting from PostgreSQL...")
    tasks = fetch_table("tasks")
    habits = fetch_table("habits")
    users = fetch_users()
    projects = fetch_projects()

    print("[ETL] Extracting from MongoDB (AI/LLM)...")
    ai_logs = fetch_collection("model_usage")
    ai_models = fetch_ai_models()
    conversations = fetch_conversations()
    cost_tracking = fetch_cost_tracking()

    print("[ETL] Extracting from notes-server MongoDB...")
    notes = fetch_notes()
    journals = fetch_journals()

    print("[ETL] Transforming and joining data...")
    combined = normalize_and_join(
        tasks, habits, ai_logs,
        users=users,
        projects=projects,
        ai_models=ai_models,
        conversations=conversations,
        cost_tracking=cost_tracking,
        notes=notes,
        journals=journals
    )

    print("[ETL] Cleaning and enriching data...")
    combined = clean_text_fields(combined, ['content'])
    combined = fill_missing(combined, {'user_id': 'unknown'})
    combined = strip_all_string_columns(combined)
    if 'email' in combined.columns:
        combined = lowercase_column(combined, 'email')
    combined = normalize_categorical(combined, ['event_type', 'source'])
    combined = add_time_features(combined)
    combined = fill_all_missing(combined)

    # Convert UUID and ObjectId columns to strings for Parquet compatibility
    combined = convert_special_types_to_str(combined)

    # Ensure output directory exists
    output_dir = os.path.join(os.path.dirname(__file__), '..', 'etl_output')
    output_dir = os.path.abspath(output_dir)
    os.makedirs(output_dir, exist_ok=True)
    duckdb_path = os.path.join(output_dir, 'analytics.duckdb')
    parquet_path = os.path.join(output_dir, 'combined.parquet')

    print("[ETL] Loading to DuckDB...")
    save_to_duckdb(combined, db_path=duckdb_path)

    print("[ETL] Loading to Parquet...")
    save_to_parquet(combined, path=parquet_path)

    print("[ETL] Pipeline complete.")


if __name__ == "__main__":
    run_pipeline()
