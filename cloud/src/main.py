from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from api.routes import router
from models.firebase import init_firebase
from config import settings

app = FastAPI(
    title="STX AI Cloud API",
    description="Financial AI Agent API — stocks, forex, crypto analysis",
    version="0.1.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(router)


@app.on_event("startup")
async def startup():
    init_firebase()


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=settings.api_port)
