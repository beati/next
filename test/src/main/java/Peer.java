package test;

import java.io.File;
import org.openqa.selenium.By;
import org.openqa.selenium.WebDriver;
import org.openqa.selenium.WebElement;
import org.openqa.selenium.chrome.ChromeOptions;
import org.openqa.selenium.chrome.ChromeDriver;
import org.openqa.selenium.support.ui.ExpectedCondition;
import org.openqa.selenium.support.ui.ExpectedConditions;
import org.openqa.selenium.support.ui.WebDriverWait;

public class Peer {
	private WebDriver driver;
	private String name;

	public Peer(String name) {
		this.name = name;
	}

	public void load() {
		String tmpDir = System.getProperty("java.io.tmpdir");
		String profileDir = tmpDir + File.separator + "chrome_profiles" + File.separator + this.name;
		System.out.println(profileDir);

		ChromeOptions options = new ChromeOptions();
		options.addArguments(
			"--user-data-dir=" + profileDir,
			"--use-fake-device-for-media-stream"
		);
		this.driver = new ChromeDriver(options);
		//this.driver.get("https://next.beati.io");
		this.driver.get("https://localhost:2001");
	}

	public void register(Boolean button) {
		WebElement input = (new WebDriverWait(this.driver, 10).until(ExpectedConditions.presenceOfElementLocated(By.id("md-input-0-input"))));
		input.sendKeys(this.name);

		if (button) {
			WebElement register = this.driver.findElement(By.id("register"));
			register.click();
		} else {
			input.submit();
		}
	}

	public void next() {
		WebElement next = (new WebDriverWait(this.driver, 10).until(ExpectedConditions.presenceOfElementLocated(By.id("next"))));
		next.click();
	}

	public Boolean isWaitingForPeer() {
		try {
			new WebDriverWait(this.driver, 10).until(ExpectedConditions.presenceOfElementLocated(By.id("waiting")));
		} catch (Exception e) {
			return false;
		}
		return true;
	}

	public Boolean isMatched() {
		try {
			new WebDriverWait(this.driver, 10).until(ExpectedConditions.presenceOfElementLocated(By.id("peerName")));
		} catch (Exception e) {
			return false;
		}
		return true;
	}

	public Boolean isLocalVideoPlaying() {
		try {
			new WebDriverWait(this.driver, 10).until(new ExpectedCondition<Boolean>() {
				public Boolean apply(WebDriver d) {
					WebElement video = d.findElement(By.id("localVideo"));
					return Double.parseDouble(video.getAttribute("currentTime")) > 0.1;
				}
			});
		} catch (Exception e) {
			return false;
		}
		return true;
	}

	public Boolean isRemoteVideoPlaying() {
		try {
			new WebDriverWait(this.driver, 10).until(new ExpectedCondition<Boolean>() {
				public Boolean apply(WebDriver d) {
					WebElement video = d.findElement(By.id("remoteVideo"));
					return Double.parseDouble(video.getAttribute("currentTime")) > 0.1;
				}
			});
		} catch (Exception e) {
			return false;
		}
		return true;
	}

	public String localSrc() {
		WebElement element = this.driver.findElement(By.id("localVideo"));
		//System.out.println(element.getAttribute("currentTime"));
		//System.out.println(element.getAttribute("played"));
		return element.getAttribute("src");
	}

	/*
	public Boolean isLocalPlaying() {
		WebElement element = this.driver.findElement(By.id("localVideo"));
	}

	public Boolean isRemotePlaying() {
		return false;
	}
	*/

	public void quit() {
		this.driver.quit();
	}

	private void sleep(long millis) {
		try {
			Thread.sleep(millis);
		} catch (InterruptedException e) {
		}
	}
}
