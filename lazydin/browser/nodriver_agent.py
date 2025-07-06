import logging
import time
from uuid import uuid4

from nodriver import Browser, Config, start, Element


class NoDriverService:
    drivers: dict[str, Browser] = {}

    def __create_options(self) -> Config:
        opts = Config(headless=False)
        return opts

    async def __create_driver(self):
        browser = await start(config=self.__create_options())
        return browser

    async def open_browser(self) -> str:
        logging.debug(f"Open browser in local")
        driver = await self.__create_driver()
        id = uuid4()
        self.drivers[str(id)] = driver

        return str(id)
    
    def close(self):
        for idx, driver in enumerate(self.drivers.values()):
            driver.stop()
            

    def __del__(self):
        self.close()

    @staticmethod
    async def human_input_simulate(element: Element, content: str, delay=1):
        for key in content:
            time.sleep(delay)
            await element.send_keys(key)
