import uvicorn
from decouple import config
from fastapi import FastAPI, Form, Request
from fastapi.concurrency import run_in_threadpool
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates
from pydantic import BaseModel

from lazydin.browser import RemoteBrowserService

templates = Jinja2Templates(directory="lazydin/template")



# Função síncrona que roda Selenium
def run_selenium(url: str) -> str:
    # TODO: preciso depois de uma pool de broserser
    browser_service = RemoteBrowserService(selenium_remote_url=config("SELENIUM_GRID_URL"))
    driver_key = browser_service.open_browser()
    with browser_service.drivers[driver_key] as driver:
        driver.get(url)
        return driver.page_source.title()


app = FastAPI()


# Modelo de dados para o corpo da requisição
class ScrapeRequest(BaseModel):
    url: str


@app.get("/health")
async def health():
    return ":-)"


# Endpoint assíncrono com corpo da requisição
# Endpoint que retorna HTML parcial
@app.post("/scrape", response_class=HTMLResponse)
async def scrape(url: str = Form(...)):
    title = await run_in_threadpool(run_selenium, url)
    return f"<p><strong>Título da página:</strong> {title}</p>"


@app.get("/", response_class=HTMLResponse)
async def home(request: Request):
    return templates.TemplateResponse("index.html", {"request": request})

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
