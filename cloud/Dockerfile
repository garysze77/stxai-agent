FROM python:3.12-slim

WORKDIR /app

COPY pyproject.toml .
RUN python -c "import tomllib; \
    deps = tomllib.load(open('pyproject.toml', 'rb'))['project']['dependencies']; \
    open('requirements.txt', 'w').write('\n'.join(deps))" && \
    pip install --no-cache-dir -r requirements.txt && \
    rm requirements.txt

COPY src/ ./src/
ENV PYTHONPATH=/app/src

EXPOSE 8000
CMD ["sh", "-c", "uvicorn main:app --host 0.0.0.0 --port ${PORT:-8000}"]
