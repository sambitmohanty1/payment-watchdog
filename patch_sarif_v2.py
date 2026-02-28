import re

def process_file(filename):
    with open(filename, 'r') as f:
        content = f.read()

    # The previous script might not have matched because the sarif_file line
    # could have trailing spaces or different indentation. We'll use a more flexible regex.

    replacements = [
        (r"(sarif_file:\s*'trivy-fs-results.sarif')", r"\1\n          category: 'trivy-fs'"),
        (r"(sarif_file:\s*'trivy-api-results.sarif')", r"\1\n          category: 'trivy-api'"),
        (r"(sarif_file:\s*'trivy-worker-results.sarif')", r"\1\n          category: 'trivy-worker'"),
        (r"(sarif_file:\s*'trivy-web-results.sarif')", r"\1\n          category: 'trivy-web'"),
        (r"(sarif_file:\s*'api/gosec-api-results.sarif')", r"\1\n          category: 'gosec-api'"),
        (r"(sarif_file:\s*'worker/gosec-worker-results.sarif')", r"\1\n          category: 'gosec-worker'")
    ]

    for pat, rep in replacements:
        content = re.sub(pat, rep, content)

    with open(filename, 'w') as f:
        f.write(content)

process_file('.github/workflows/security-scan.yml')
process_file('.github/workflows/payment-watchdog-ci.yml')
