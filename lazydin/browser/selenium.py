import logging
import time

import fake_useragent
from selenium import webdriver
from selenium.webdriver.remote.webdriver import WebDriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.remote.webelement import WebElement

class RemoteBrowserService:
    drivers:  dict[str, WebDriver]
    selenium_remote_url: str
    os: str
    browsers: list[str]
    maximize_window: bool

    def __init__(
        self,
        selenium_remote_url: str,
        os="windows",
        browsers=["chrome"],
        maximize_window=True,
    ) -> None:
        self.drivers: dict[str, WebDriver] = {}
        self.selenium_remote_url = selenium_remote_url
        self.os = os
        self.browsers = browsers
        self.maximize_window = maximize_window

    def __create_options(self) -> Options:
        opts = Options()
        opts.add_argument(
            f"user-agent={fake_useragent.UserAgent(os=self.os, browsers=self.browsers)}"
        )
        return opts
    def __create_driver(self) -> WebDriver:
        driver = webdriver.Remote(
            options=self.__create_options(), command_executor=self.selenium_remote_url
        )
        return driver

    def open_browser(self) -> str:
        logging.debug(f"Open browser in {self.selenium_remote_url}")
        driver = self.__create_driver()

        if not driver.session_id:
            raise Exception("Not able to regisrer a driver")
        if self.maximize_window:
            driver.maximize_window()
        self.drivers[driver.session_id] = driver

        return driver.session_id

    def close(self, driver_key: str):
        self.drivers[driver_key].close()
        del self.drivers[driver_key]

    def __del__(self):
        for session_id, driver in self.drivers.items():
            try:
                driver.quit()
                time.sleep(1)
                logging.debug(f"Closed driver with session_id {session_id}")
            except Exception as e:
                logging.error(f"Error when close driver {session_id}: {e}")
            finally:
                try:
                    if session_id:
                        logging.error(
                            f"Derrubando o driver com session_id {session_id} do Selenium Grid."
                        )
                        webdriver.Remote(
                            options=self.__create_options(),
                            command_executor=driver.command_executor,
                        ).quit()
                except Exception as e:
                    logging.error(
                        f"Erro ao desconectar o driver com session_id {session_id} do Selenium Grid: {e}"
                    )
    @staticmethod
    def human_input_simulate(element: WebElement, content: str, delay=1):
        for key in content:
            time.sleep(delay)
            element.send_keys(key)

