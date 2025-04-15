import pandas as pd
import numpy as np


def flatten_dict_column(df, col):
    """Flatten a column of dicts in a DataFrame, joining keys as col_key."""
    if col in df.columns:
        dict_df = df[col].apply(pd.Series)
        dict_df = dict_df.add_prefix(f"{col}_")
        df = pd.concat([df.drop(columns=[col]), dict_df], axis=1)
    return df


def flatten_list_column(df, col):
    """Flatten a column of lists in a DataFrame, joining values as comma-separated string."""
    if col in df.columns:
        df[col] = df[col].apply(lambda x: ','.join(
            map(str, x)) if isinstance(x, list) else x)
    return df


def normalize_and_join(
    tasks=None, habits=None, ai_logs=None, users=None, projects=None, ai_models=None, conversations=None, cost_tracking=None, notes=None, journals=None
):
    """
    Normalize and join all sources into a wide analytics DataFrame, retaining as many relevant fields as possible.
    Explicitly map timestamp and user_id for all sources.
    """
    dfs = []
    # Tasks
    if tasks is not None and not tasks.empty:
        t = tasks.copy()
        t['source'] = 'task'
        t['event_type'] = 'task_event'
        if 'created_at' in t.columns:
            t['timestamp'] = pd.to_datetime(t['created_at'], errors='coerce')
        elif 'createdAt' in t.columns:
            t['timestamp'] = pd.to_datetime(t['createdAt'], errors='coerce')
        else:
            t['timestamp'] = pd.NaT
        if 'user_id' not in t.columns:
            t['user_id'] = 'unknown'
        dfs.append(t)
    # Habits
    if habits is not None and not habits.empty:
        h = habits.copy()
        h['source'] = 'habit'
        h['event_type'] = 'habit_event'
        if 'created_at' in h.columns:
            h['timestamp'] = pd.to_datetime(h['created_at'], errors='coerce')
        elif 'createdAt' in h.columns:
            h['timestamp'] = pd.to_datetime(h['createdAt'], errors='coerce')
        else:
            h['timestamp'] = pd.NaT
        if 'user_id' not in h.columns:
            h['user_id'] = 'unknown'
        dfs.append(h)
    # AI logs/model usage
    if ai_logs is not None and not ai_logs.empty:
        a = ai_logs.copy()
        a['source'] = 'ai_log'
        a['event_type'] = a.get('request_type', 'ai_event')
        # Robust timestamp mapping
        if 'timestamp' in a.columns:
            a['timestamp'] = pd.to_datetime(a['timestamp'], errors='coerce')
        elif 'created_at' in a.columns:
            a['timestamp'] = pd.to_datetime(a['created_at'], errors='coerce')
        elif 'createdAt' in a.columns:
            a['timestamp'] = pd.to_datetime(a['createdAt'], errors='coerce')
        else:
            a['timestamp'] = pd.NaT
        if 'user_id' not in a.columns:
            a['user_id'] = 'unknown'
        dfs.append(a)
    # Users
    if users is not None and not users.empty:
        u = users.copy()
        u['source'] = 'user'
        u['event_type'] = 'user_event'
        if 'created_at' in u.columns:
            u['timestamp'] = pd.to_datetime(u['created_at'], errors='coerce')
        elif 'createdAt' in u.columns:
            u['timestamp'] = pd.to_datetime(u['createdAt'], errors='coerce')
        else:
            u['timestamp'] = pd.NaT
        if 'id' in u.columns:
            u['user_id'] = u['id']
        elif 'user_id' not in u.columns:
            u['user_id'] = 'unknown'
        dfs.append(u)
    # Projects
    if projects is not None and not projects.empty:
        p = projects.copy()
        p['source'] = 'project'
        p['event_type'] = 'project_event'
        if 'created_at' in p.columns:
            p['timestamp'] = pd.to_datetime(p['created_at'], errors='coerce')
        elif 'createdAt' in p.columns:
            p['timestamp'] = pd.to_datetime(p['createdAt'], errors='coerce')
        else:
            p['timestamp'] = pd.NaT
        if 'creator_id' in p.columns:
            p['user_id'] = p['creator_id']
        elif 'user_id' not in p.columns:
            p['user_id'] = 'unknown'
        dfs.append(p)
    # AI Models
    if ai_models is not None and not ai_models.empty:
        m = ai_models.copy()
        m['source'] = 'ai_model'
        m['event_type'] = 'ai_model_event'
        if 'created_at' in m.columns:
            m['timestamp'] = pd.to_datetime(m['created_at'], errors='coerce')
        elif 'createdAt' in m.columns:
            m['timestamp'] = pd.to_datetime(m['createdAt'], errors='coerce')
        else:
            m['timestamp'] = pd.NaT
        if 'user_id' not in m.columns:
            m['user_id'] = 'unknown'
        dfs.append(m)
    # Conversations
    if conversations is not None and not conversations.empty:
        c = conversations.copy()
        c['source'] = 'conversation'
        c['event_type'] = 'conversation_event'
        if 'last_message_time' in c.columns:
            c['timestamp'] = pd.to_datetime(
                c['last_message_time'], errors='coerce')
        elif 'created_at' in c.columns:
            c['timestamp'] = pd.to_datetime(c['created_at'], errors='coerce')
        elif 'createdAt' in c.columns:
            c['timestamp'] = pd.to_datetime(c['createdAt'], errors='coerce')
        else:
            c['timestamp'] = pd.NaT
        if 'user_id' not in c.columns:
            c['user_id'] = 'unknown'
        dfs.append(c)
    # Cost Tracking
    if cost_tracking is not None and not cost_tracking.empty:
        ct = cost_tracking.copy()
        ct['source'] = 'cost_tracking'
        ct['event_type'] = 'cost_event'
        if 'timestamp' in ct.columns:
            ct['timestamp'] = pd.to_datetime(ct['timestamp'], errors='coerce')
        elif 'created_at' in ct.columns:
            ct['timestamp'] = pd.to_datetime(ct['created_at'], errors='coerce')
        elif 'createdAt' in ct.columns:
            ct['timestamp'] = pd.to_datetime(ct['createdAt'], errors='coerce')
        else:
            ct['timestamp'] = pd.NaT
        if 'user_id' not in ct.columns:
            ct['user_id'] = 'unknown'
        dfs.append(ct)
    # Notes
    if notes is not None and not notes.empty:
        n = notes.copy()
        n['source'] = 'note'
        n['event_type'] = 'note_event'
        n = flatten_list_column(n, 'tags')
        n = flatten_list_column(n, 'checklist')
        if 'createdAt' in n.columns:
            n['timestamp'] = pd.to_datetime(n['createdAt'], errors='coerce')
        elif 'created_at' in n.columns:
            n['timestamp'] = pd.to_datetime(n['created_at'], errors='coerce')
        else:
            n['timestamp'] = pd.NaT
        if 'userId' in n.columns:
            n['user_id'] = n['userId']
        elif 'user_id' not in n.columns:
            n['user_id'] = 'unknown'
        dfs.append(n)
    # Journals
    if journals is not None and not journals.empty:
        j = journals.copy()
        j['source'] = 'journal'
        j['event_type'] = 'journal_event'
        if 'createdAt' in j.columns:
            j['timestamp'] = pd.to_datetime(j['createdAt'], errors='coerce')
        elif 'created_at' in j.columns:
            j['timestamp'] = pd.to_datetime(j['created_at'], errors='coerce')
        else:
            j['timestamp'] = pd.NaT
        if 'userId' in j.columns:
            j['user_id'] = j['userId']
        elif 'user_id' not in j.columns:
            j['user_id'] = 'unknown'
        dfs.append(j)
    # Combine all
    if dfs:
        combined = pd.concat(dfs, ignore_index=True, sort=False)
        # Try to flatten any dict columns
        for col in combined.columns:
            if combined[col].apply(lambda x: isinstance(x, dict)).any():
                combined = flatten_dict_column(combined, col)
        # Ensure all timestamps are UTC and tz-naive
        combined['timestamp'] = pd.to_datetime(
            combined['timestamp'], utc=True, errors='coerce')
        combined['timestamp'] = combined['timestamp'].dt.tz_localize(None)
        combined = combined.sort_values('timestamp')
        return combined
    else:
        return pd.DataFrame()
