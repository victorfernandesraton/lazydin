import logging
from uuid import uuid4

from nodriver import Browser, Config, start, Element
from asyncio import sleep

logger = logging.getLogger(__name__)


class NoDriverService:
    drivers: dict[str, Browser] = {}
    keep_browser_open: bool = True

    def __create_options(self) -> Config:
        opts = Config(headless=False)
        return opts

    async def __create_driver(self):
        browser = await start(config=self.__create_options())
        return browser

    async def open_browser(self) -> str:
        driver = await self.__create_driver()
        id = uuid4()
        self.drivers[str(id)] = driver

        logger.debug(f"Open browser in local with driver {id}")

        return str(id)

    def close(self):
        if not self.keep_browser_open:
            return

        for idx, driver in enumerate(self.drivers.values()):
            driver.stop()

    def __del__(self):
        self.close()

    @staticmethod
    async def human_input_simulate(element: Element, content: str, delay=1):
        for key in content:
            await sleep(delay)
            await element.send_keys(key)
