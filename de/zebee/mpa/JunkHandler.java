package de.zebee.mpa;

import java.io.IOException;

/**
 * @author Sebastian Gesemann
 */
public interface JunkHandler {

	public void write(int bite) throws IOException;
	
	public void endOfJunkBlock() throws IOException;

}
