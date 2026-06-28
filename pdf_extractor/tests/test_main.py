from fastapi.testclient import TestClient
from main import app

client = TestClient(app)
SAMPLE_PDF = "tests/testdata/sample.pdf"
EMPTY_PDF = "tests/testdata/empty.pdf"

def test_extract_pdf_success():
    with open(SAMPLE_PDF, "rb") as f:
        response = client.post(
            "/extract",
            files={"file": ("sample.pdf", f, "application/pdf")},
        )
        
    assert response.status_code == 200
    body = response.json()
    assert "text" in body
    assert isinstance(body["text"], str)
    assert len(body["text"]) > 0
    assert "Hello from test PDF" in response.json()["text"]
    
def test_extract_rejects_non_pdf():
    with open("tests/testdata/not_a_pdf.txt", "rb") as f:
        response = client.post(
            "/extract",
            files={"file": ("doc.txt", f, "text/plain")},
        )

    assert response.status_code == 415
    
def test_extract_rejects_pdf_with_wrong_mime():
    with open(SAMPLE_PDF, "rb") as f:
        response = client.post(
            "/extract",
            files={"file": ("sample.pdf", f, "application/octet-stream")},
        )

    assert response.status_code == 415
    
def test_extract_empty_pdf():
    with open(EMPTY_PDF, "rb") as f:
        response = client.post(
            "/extract",
            files={"file": ("empty.pdf", f, "application/pdf")},
        )
        
    assert response.status_code == 200
    assert response.json()["text"] == ""
    
def test_extract_missing_file():
    response = client.post("/extract")
    assert response.status_code == 422
    
def test_extract_corrupted_pdf():
    response = client.post(
        "/extract",
        files={"file": ("broken.pdf", b"not a real pdf", "application/pdf")},
    )

    assert response.status_code == 400
    
def test_extract_invalid_pdf_signature():
    response = client.post(
        "/extract",
        files={
            "file": (
                "fake.pdf",
                b"this is not a pdf",
                "application/pdf",
            )
        },
    )

    assert response.status_code == 400