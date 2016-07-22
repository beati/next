package test;

public class Test  {
	public static void main(String[] args) {
		Peer alice = new Peer("Alice");
		alice.load();

		Peer bob = new Peer("Bob");
		bob.load();

		Peer carole = new Peer("Carole");
		carole.load();

		System.out.println();

		alice.register(false);
		if (!alice.isWaitingForPeer() || !alice.isLocalVideoPlaying()) {
			System.out.println("Test fail");
			System.exit(1);
		}
		System.out.println("Test pass");

		bob.register(true);
		if (!alice.isMatched() || !bob.isMatched() ||
			!alice.isRemoteVideoPlaying() || !bob.isRemoteVideoPlaying()) {
			System.out.println("Test fail");
			System.exit(1);
		}
		System.out.println("Test pass");

		carole.register(true);
		if (!carole.isWaitingForPeer()) {
			System.out.println("Test fail");
			System.exit(1);
		}
		System.out.println("Test pass");

		bob.next();
		if (!alice.isWaitingForPeer() || !bob.isMatched() || !carole.isMatched()) {
			System.out.println("Test fail");
			System.exit(1);
		}
		System.out.println("Test pass");

		System.out.println("\nPress enter to exit");
		try {
			System.in.read();
		} catch (Exception e) {
		}

		alice.quit();
		bob.quit();
		carole.quit();

		/*
		(new WebDriverWait(driver, 10)).until(new ExpectedCondition<Boolean>() {
			public Boolean apply(WebDriver d) {
				return d.getTitle().toLowerCase().startsWith("cheese!");
			}
		});
		*/

		//driver.quit();
	}
}
