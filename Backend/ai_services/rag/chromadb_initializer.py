# chromadb_initializer.py

import chromadb
from sentence_transformers import SentenceTransformer
import uuid
import re
import os
from PyPDF2 import PdfReader
import glob

print("Starting ChromaDB initialization...")

# Get the directory where this script is located
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
KNOWLEDGE_BASE_DIR = os.path.join(SCRIPT_DIR, "knowledge_base")

def clean_text(text):
    """Clean and fix common PDF extraction issues"""
    # First, let's separate words that are incorrectly joined
    text = re.sub(r'(?<=[a-z])(?=[A-Z])', ' ', text)  # Add space between camelCase
    text = re.sub(r'(?<=[A-Za-z])(?=\d)|(?<=\d)(?=[A-Za-z])', ' ', text)  # Add space between letters and numbers
    
    # Fix common word splits
    common_fixes = {
        'comp ass': 'compass',
        'w orks': 'works',
        'c ompass': 'compass',
        't ask': 'task',
        'g uide': 'guide',
        'u ser': 'user',
        'a i': 'ai'
    }
    
    for wrong, correct in common_fixes.items():
        text = re.sub(rf'\b{wrong}\b', correct, text, flags=re.IGNORECASE)
    
    # Fix spacing issues
    text = re.sub(r'\s+', ' ', text)  # Replace multiple spaces with single space
    text = re.sub(r'\s*([.,!?;:])\s*', r'\1 ', text)  # Fix spacing around punctuation
    text = re.sub(r'\s*\n\s*', '\n', text)  # Fix newline spacing
    
    # Add proper spacing after periods if missing
    text = re.sub(r'\.(?=[A-Z])', '. ', text)
    
    # Fix parentheses spacing
    text = re.sub(r'\s*\(\s*', ' (', text)
    text = re.sub(r'\s*\)\s*', ') ', text)
    
    # Ensure proper spacing around special characters
    text = re.sub(r'(?<=\w)[-/](?=\w)', ' - ', text)  # Add spaces around hyphens between words
    
    return text.strip()

def extract_text_from_pdf(pdf_path):
    print(f"Processing PDF: {os.path.basename(pdf_path)}")
    try:
        reader = PdfReader(pdf_path)
        text = ""
        for page in reader.pages:
            page_text = page.extract_text()
            # Clean text as soon as we extract it from each page
            page_text = clean_text(page_text)
            text += page_text + "\n\n"  # Add double newline between pages for better separation
        return text
    except Exception as e:
        print(f"Error reading PDF {pdf_path}: {str(e)}")
        return None

def chunk_text(text, max_chunk_size=1000):
    """Split text into smaller chunks for better embedding"""
    # First clean any remaining issues in the full text
    text = clean_text(text)
    
    # Split on sentences while preserving them
    # Look for sentence endings followed by spaces and capital letters
    sentences = re.split(r'(?<=[.!?])\s+(?=[A-Z])', text)
    chunks = []
    current_chunk = ""
    
    for sentence in sentences:
        sentence = sentence.strip()
        if not sentence:
            continue
            
        if len(current_chunk) + len(sentence) < max_chunk_size:
            current_chunk += sentence + " "
        else:
            if current_chunk:
                chunks.append(current_chunk.strip())
            current_chunk = sentence + " "
    
    if current_chunk:
        chunks.append(current_chunk.strip())
    
    return chunks

print("Creating ChromaDB client...")
try:
    # Initialize ChromaDB (new API)
    chroma_store_path = os.path.join(SCRIPT_DIR, "chroma_store")
    chroma_client = chromadb.PersistentClient(path=chroma_store_path)
    print("ChromaDB client created successfully")
except Exception as e:
    print(f"Error creating ChromaDB client: {str(e)}")
    raise

# Create or get a collection
collection_name = "knowledge_base"
try:
    if collection_name in [c.name for c in chroma_client.list_collections()]:
        print(f"Getting existing collection: {collection_name}")
        collection = chroma_client.get_collection(collection_name)
    else:
        print(f"Creating new collection: {collection_name}")
        collection = chroma_client.create_collection(collection_name)
except Exception as e:
    print(f"Error with collection: {str(e)}")
    raise

print("Initializing sentence transformer...")
try:
    # Embedder
    embedder = SentenceTransformer("all-MiniLM-L6-v2")
    print("Sentence transformer initialized successfully")
except Exception as e:
    print(f"Error initializing sentence transformer: {str(e)}")
    raise

print("Processing PDF files from knowledge base...")
try:
    # Get all PDF files in the knowledge base directory
    pdf_files = glob.glob(os.path.join(KNOWLEDGE_BASE_DIR, "*.pdf"))
    total_chunks_added = 0
    
    for pdf_file in pdf_files:
        file_name = os.path.basename(pdf_file)
        print(f"\nProcessing {file_name}...")
        
        # Extract text from PDF
        text = extract_text_from_pdf(pdf_file)
        if text is None:
            continue
            
        # Split text into chunks
        chunks = chunk_text(text)
        print(f"Created {len(chunks)} chunks from {file_name}")
        
        # Add chunks to ChromaDB
        for i, chunk in enumerate(chunks):
            embedding = embedder.encode(chunk).tolist()
            collection.add(
                documents=[chunk],
                embeddings=[embedding],
                ids=[f"{file_name}_{i}_{str(uuid.uuid4())}"],
                metadatas=[{
                    "source": file_name,
                    "chunk_index": i,
                    "total_chunks": len(chunks)
                }]
            )
            if (i + 1) % 5 == 0:
                print(f"Added {i + 1}/{len(chunks)} chunks from {file_name}...")
        
        total_chunks_added += len(chunks)
        print(f"✅ Completed processing {file_name}")
    
    print(f"\n✅ Successfully added {total_chunks_added} chunks from {len(pdf_files)} PDF files to ChromaDB.")
except Exception as e:
    print(f"Error processing PDF files: {str(e)}")
    raise