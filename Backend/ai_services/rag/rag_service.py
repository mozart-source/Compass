import chromadb
from sentence_transformers import SentenceTransformer
import os


class RAGService:
    def __init__(self, embedder=None):
        self.script_dir = os.path.dirname(os.path.abspath(__file__))

        # Use environment variable for chroma store path if available
        self.chroma_store_path = os.environ.get("CHROMA_STORE_PATH",
                                                os.path.join(self.script_dir, "chroma_store"))

        # Use environment variable for sentence transformer cache if available
        cache_path = os.environ.get(
            "SENTENCE_TRANSFORMERS_HOME", "/app/cache/sentence-transformers")
        os.makedirs(cache_path, exist_ok=True)

        # Initialize the sentence transformer with the cache path
        self.embedder = embedder if embedder else SentenceTransformer(
            "all-MiniLM-L6-v2", cache_folder=cache_path)

        # Ensure the chroma store directory exists
        os.makedirs(self.chroma_store_path, exist_ok=True)

        try:
            self.chroma_client = chromadb.PersistentClient(
                path=self.chroma_store_path)
            # Try to get or create the collection
            try:
                self.collection = self.chroma_client.get_collection(
                    "knowledge_base")
            except Exception:
                # Create collection if it doesn't exist
                self.collection = self.chroma_client.create_collection(
                    "knowledge_base")
        except Exception as e:
            print(f"Error initializing ChromaDB: {str(e)}")
            # Fallback to empty collection
            self.chroma_client = None
            self.collection = None

    async def get_relevant_context(self, query: str, n_results: int = 3) -> str:
        """
        Get relevant context from ChromaDB based on the query.

        Args:
            query: The user's query
            n_results: Number of relevant chunks to retrieve

        Returns:
            A string containing the relevant context
        """
        # Check if ChromaDB is available
        if self.collection is None:
            return ""

        try:
            # Generate embedding for the query
            query_embedding = self.embedder.encode(query).tolist()

            # Query the collection
            results = self.collection.query(
                query_embeddings=[query_embedding],
                n_results=n_results,
                include=["documents", "metadatas"]
            )

            # Format the results into a context string
            context_chunks = []
            for doc, metadata in zip(results['documents'][0], results['metadatas'][0]):
                source = metadata.get('source', 'Unknown')
                context_chunks.append(f"From {source}:\n{doc}\n")

            return "\n".join(context_chunks)
        except Exception as e:
            print(f"Error getting context: {str(e)}")
            return ""

    def get_sources_used(self, query: str, n_results: int = 3) -> list:
        """
        Get the sources of the documents used for context.

        Args:
            query: The user's query
            n_results: Number of relevant chunks to retrieve

        Returns:
            List of source filenames used
        """
        if self.collection is None:
            return []

        try:
            query_embedding = self.embedder.encode(query).tolist()
            results = self.collection.query(
                query_embeddings=[query_embedding],
                n_results=n_results,
                include=["metadatas"]
            )

            sources = [meta.get('source', 'Unknown')
                       for meta in results['metadatas'][0]]
            return list(set(sources))  # Remove duplicates
        except Exception as e:
            print(f"Error getting sources: {str(e)}")
            return []
