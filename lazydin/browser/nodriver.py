from nodriver import Config
import nodriver
class NoDriverService:
    drivers:  dict[str, WebDriver]
    selenium_remote_url: str
    os: str
    browsers: list[str]
    maximize_window: bool
    def __create_options(self) -> n:
        opts = Config(headless=False)
        return opts

    def __create_driver(self):
        browser = nodriver.

    def open_browser(self) -> str:
        logging.debug(f"Open browser in {self.selenium_remote_url}")
        driver = self.__create_driver()

        if not driver.session_id:
            raise Exception("Not able to regisrer a driver")
        if self.maximize_window:
            driver.maximize_window()
        self.drivers[driver.session_id] = driver

        return driver.session_id


