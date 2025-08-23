import logging
import nodriver
from nodriver import Config, Browser

class NoDriverService:
    drivers: dict[str, Browser] = {}
    selenium_remote_url: str = None
    os: str = "linux"
    browsers: list[str] = ["chrome"]
    maximize_window: bool = True
    cookie_file: str = ".session.dat"
    
    def __init__(
        self,
        selenium_remote_url: str = None,
        os: str = "linux",
        browsers: list[str] = ["chrome"],
        maximize_window: bool = True,
        cookie_file: str = ".session.dat"
    ):
        self.selenium_remote_url = selenium_remote_url
        self.os = os
        self.browsers = browsers
        self.maximize_window = maximize_window
        self.cookie_file = cookie_file
        self.drivers = {}
    
    def __create_options(self) -> Config:
        opts = Config(headless=False)
        return opts

    async def __create_driver(self):
        config = self.__create_options()
        browser = await Browser.create(config)
        return browser

    async def open_browser(self) -> str:
        logging.debug(f"Opening browser")
        driver = await self.__create_driver()

        # Use the browser's ID as the session_id
        session_id = str(id(driver))
        self.drivers[session_id] = driver

        return session_id
        
    # Cookie methods removed - we'll use browser.cookies directly
    
    def close(self, driver_key: str = None):
        """Close a specific browser or all browsers"""
        if driver_key:
            if driver_key in self.drivers:
                browser = self.drivers[driver_key]
                browser.stop()
                del self.drivers[driver_key]
                logging.info(f"Closed browser with key {driver_key}")
            else:
                logging.warning(f"Driver key {driver_key} not found")
        else:
            # Close all browsers
            for key, browser in list(self.drivers.items()):
                browser.stop()
                del self.drivers[key]
            logging.info("Closed all browsers")
    
    def __del__(self):
        """Cleanup when the service is garbage collected"""
        for browser in self.drivers.values():
            try:
                browser.stop()
            except:
                pass
        self.drivers.clear()
        
    @staticmethod
    async def human_input_simulate(element, content: str, delay=1):
        """Simulate human typing with random delays between keystrokes"""
        import random
        import asyncio
        
        for char in content:
            await element.send(char)
            # Random delay between keystrokes
            await asyncio.sleep(delay * random.uniform(0.5, 1.5))


