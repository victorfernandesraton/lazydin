import logging

from lazydin.browser import NoDriverService


class LinkedinAuth:
    USERNAME_INPUT_XPATH = "//input[@id='username']"
    PASSWORD_INPUT_XPATH = "//input[@id='password']"
    BUTTON_SUBMIT_XPATH = "//button[@type='submit']"
    browser_service: NoDriverService
    linkedin_domain: str

    def __init__(
        self,
        browser_service: NoDriverService,
        linkedin_domain="https://www.linkedin.com",
    ):
        self.browser_service = browser_service
        self.linkedin_domain = linkedin_domain

    async def execute(self, driver_key: str, username: str, password: str):
        logging.info("go to site")
        driver = self.browser_service.drivers[driver_key]
        page = await driver.get(f"{self.linkedin_domain}/login")
        await page.maximize()
        await page.find("Sign in", best_match=True)
        logging.info("set input")
        input_list = {
            self.USERNAME_INPUT_XPATH: username,
            self.PASSWORD_INPUT_XPATH: password,
        }
        for xpath, value in input_list.items():
            txt_input = await page.xpath(xpath,timeout=10)

            if not txt_input:
                raise Exception(f"Not found {xpath} input")

            await self.browser_service.human_input_simulate(txt_input[0], value)

        btn_submit = await page.xpath(
            xpath=self.BUTTON_SUBMIT_XPATH,
            timeout=10
        )
        if not btn_submit:
            raise Exception("Not found button")

        await btn_submit[0].click()
        logging.info("button clicked")
        await page.close()

