import io

from fastapi import FastAPI, File, HTTPException, UploadFile
from fastapi.concurrency import run_in_threadpool
from fastapi.responses import JSONResponse
import pdfplumber

MAX_SIZE = 20 * 1024 * 1024

app = FastAPI()

def extract_pdf(content: bytes) -> str:
    with pdfplumber.open(io.BytesIO(content)) as pdf:
        pages = []
        for page in pdf.pages:
            text = page.extract_text()
            if text and text.strip():
                pages.append(text.strip())
        return "\n\n".join(pages)

@app.post("/extract")
async def extract(file: UploadFile = File(...)):
    if file.content_type != "application/pdf":
        raise HTTPException(status_code=415, detail="this service only handles PDF files")
    
    content = await file.read()

    if len(content) > MAX_SIZE:
        raise HTTPException(413, "PDF exceeds maximum size.")
    
    if not content.startswith(b"%PDF"):
        raise HTTPException(status_code=400, detail="Invalid PDF.")

    try:
        text = await run_in_threadpool(extract_pdf, content)
        return JSONResponse({"text": text})
    except Exception:
        raise HTTPException(status_code=500, detail="Failed to extract PDF.")


