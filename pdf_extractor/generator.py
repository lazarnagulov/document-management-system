from docx import Document
doc = Document()
doc.add_paragraph('Hello from test DOCX')
doc.add_paragraph('Second paragraph')
doc.save('./tests/testdata/sample.docx')