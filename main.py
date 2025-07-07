import uvicorn
from decouple import config
from fastapi import FastAPI, Form, Request, BackgroundTasks
from fastapi.concurrency import run_in_threadpool
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates
from pydantic import BaseModel
from lazydin.browser import NoDriverService
from lazydin.workflows.linkedin.auth import LinkedinAuth
import nodriver as nc
import asyncio
import time

templates = Jinja2Templates(directory="lazydin/template")


# Função síncrona que roda Selenium
async def run_selenium(url: str) -> str:
    service = NoDriverService()
    driver = await service.open_browser()
    auth = LinkedinAuth(service)
    await auth.execute(driver_key=driver, username="test@gmail.com", password="test")

    del service

    return "ok"


app = FastAPI()


# Modelo de dados para o corpo da requisição
class ScrapeRequest(BaseModel):
    url: str


def run_selenium_background(url: str):
    asyncio.run(run_selenium(url))


@app.get("/health")
async def health():
    return ":-)"


# Endpoint assíncrono com corpo da requisição
# Endpoint que retorna HTML parcial
@app.post("/scrape", response_class=HTMLResponse)
async def scrape(backgroud_tasks: BackgroundTasks, url: str = Form(...)):
    backgroud_tasks.add_task(run_selenium_background, url)
    return f"<p><strong>Título da página:</strong> {url}</p>"


@ app.get("/", response_class=HTMLResponse)
async def home(request: Request):
    return templates.TemplateResponse("index.html", {"request": request})

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
