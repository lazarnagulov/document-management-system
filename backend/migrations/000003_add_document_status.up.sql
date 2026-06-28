CREATE TYPE document_status AS ENUM ('UPLOADED', 'PROCESSING', 'READY', 'FAILED');

ALTER TABLE documents 
    ADD COLUMN status document_status NOT NULL DEFAULT 'UPLOADED';