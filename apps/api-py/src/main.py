from fastapi import FastAPI

app = FastAPI()

@app.get("/")
def read_root():
    return {"message": "Python API is running!"}

@app.get("/api/health")
def health_check():
    return {"status": "ok"}
