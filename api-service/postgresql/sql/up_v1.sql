CREATE TABLE IF NOT EXISTS omnibus (
		upc VARCHAR(13) NOT NULL,
		name TEXT NOT NULL,
		price REAL NOT NULL,
		version VARCHAR(20) NOT NULL,
		pagecount INT NOT NULL,
		datecreated DATE,
		publisher TEXT NOT NULL,
		imgpath TEXT,
		isturl TEXT,
		amazonurl TEXT,
		lastupdated TIMESTAMP,
		status TEXT,
        PRIMARY KEY (upc)  
	);


	CREATE TABLE IF NOT EXISTS sale (
    date DATE,                     -- Store the date only (no time)
    upc VARCHAR(13),               -- Universal Product Code
    sale REAL NOT NULL,            -- Sale price
    platform TEXT NOT NULL,        -- Platform (e.g., IST, Amazon)
    percent INT NOT NULL,          -- Discount percentage
    PRIMARY KEY (date, upc, platform)    -- Composite primary key
	);

