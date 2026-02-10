-- Domains that are allowed to receive traffic
CREATE TABLE IF NOT EXISTS domains (
    hostname TEXT PRIMARY KEY,
    status TEXT NOT NULL CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ DEFAULT now()
);


-- Redirect rules
CREATE TABLE IF NOT EXISTS links (
    domain_hostname TEXT NOT NULL REFERENCES domains(hostname) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    destination_url TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (domain_hostname, slug)
);

-- Performance index (optional but recommended)
CREATE INDEX IF NOT EXISTS idx_links_lookup
ON links (domain_hostname, slug);
