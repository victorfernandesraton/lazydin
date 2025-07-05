import logging
import time

from selenium.common import exceptions
from selenium.webdriver.common.by import By
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.wait import WebDriverWait
from lazydin.browser import RemoteBrowserService


class LinkedinAuth:
    USERNAME_INPUT_XPATH = "//input[@id='username']"
    PASSWORD_INPUT_XPATH = "//input[@id='password']"
    BUTTON_SUBMIT_XPATH = "//*[@id='main-content']/section/div/div/form/div/button"
    browser_service: RemoteBrowserService
    linkedin_domain: str

    def __init__(self, browser_service: RemoteBrowserService, linkedin_domain="https://www.linkedin.com"):
        self.browser_service = browser_service
        self.linkedin_domain = linkedin_domain

    def execute(self, driver_key: str, username: str, password: str):
        logging.info("go to site")
        driver = self.browser_service.drivers[driver_key]
        driver.maximize_window()
        driver.get("/login")
        input_wait = WebDriverWait(self.browser_service.drivers[driver_key], timeout=30)
        logging.info("set input")
        input_list = {
            self.USERNAME_INPUT_XPATH: username,
            self.PASSWORD_INPUT_XPATH: password,
        }
        for xpath, value in input_list.items():
            try:
                txt_input = input_wait.until(
                    EC.presence_of_element_located((By.XPATH, xpath))
                )
                self.browser_service.human_input_simulate(txt_input, value)
            except exceptions.TimeoutException:
                raise Exception(f"Not found {xpath} input")

        try:
            time.sleep(5)
            btn_submit = input_wait.until(
                EC.presence_of_element_located((By.XPATH, self.BUTTON_SUBMIT_XPATH))
            )

            btn_submit.click()
            logging.info("button clicked")
        except exceptions.TimeoutException:
            raise Exception("Not found button input")
