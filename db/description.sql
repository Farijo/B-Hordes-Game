USE hordes_challenge;

DROP TABLE IF EXISTS invitation;
DROP TABLE IF EXISTS success;
DROP TABLE IF EXISTS milestone;
DROP TABLE IF EXISTS validator;
DROP TABLE IF EXISTS goal;
DROP TABLE IF EXISTS participant;
DROP TABLE IF EXISTS challenge;
DROP TABLE IF EXISTS user;

CREATE TABLE user (
    id INT PRIMARY KEY,
    name VARCHAR(30) NOT NULL,
    simplified_name VARCHAR(30) NOT NULL,
    avatar VARCHAR(255)
);
CREATE TABLE challenge (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(64) NOT NULL,
    creator INT NOT NULL,
    flags TINYINT UNSIGNED NOT NULL, -- & 0x03 = participation (open, request, invite), & 0x04 = private, & 0x08 = no api validation, & 0x30 = status (draft, review, invite)
    start_date DATETIME,
    end_date DATETIME,
    FOREIGN KEY(creator) REFERENCES user(id)
);
CREATE TABLE participant (
    user INT NOT NULL,
    challenge INTEGER NOT NULL,
    PRIMARY KEY(user,challenge),
    FOREIGN KEY(challenge) REFERENCES challenge(id)
);
CREATE TABLE goal (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    challenge INTEGER NOT NULL,
    typ TINYINT NOT NULL,
    entity SMALLINT UNSIGNED NOT NULL,
    amount INT UNSIGNED,
    x TINYINT,
    y TINYINT,
    custom VARCHAR(128),
    FOREIGN KEY(challenge) REFERENCES challenge(id)
);
CREATE TABLE validator (
    user INT NOT NULL,
    challenge INTEGER NOT NULL,
    archived BOOLEAN NOT NULL,
    PRIMARY KEY(user,challenge),
    FOREIGN KEY(challenge) REFERENCES challenge(id)
);
CREATE TABLE milestone (
    user INT NOT NULL,
    dt DATETIME(2) NOT NULL,
    isGhost BOOLEAN,
    playedMaps INT,
    rewards BLOB, -- BLOB = concatenation de uint16, uint32 (id du picto, nombre de picto)
    dead BOOLEAN,
    isOut BOOLEAN,
    ban BOOLEAN,
    baseDef TINYINT,
    x TINYINT,
    y TINYINT,
    job TINYINT,
    mapWid TINYINT,
    mapHei TINYINT,
    mapDays TINYINT,
    conspiracy BOOLEAN,
    guide INT,
    shaman INT,
    custom BOOLEAN,
    door BOOLEAN,
    cityWater INT,
    chaos BOOLEAN,
    devast BOOLEAN,
    hard BOOLEAN,
    cityX TINYINT,
    cityY TINYINT,
    buildings BLOB,
    z INT,
    def INT,
    water INT,
    regenDir TINYINT,
    total INT,
    base INT,
    defBuildings INT,
    defUpgrades INT,
    items INT,
    itemsMul FLOAT,
    citizenHomes INT,
    citizenGuardians INT,
    watchmen INT,
    souls INT,
    temp INT,
    cadavers INT,
    bonus FLOAT,
    upgrades BLOB,
    estiMin INT,
    estiMax INT,
    estiMaxed BOOLEAN,
    nextMin INT,
    nextMax INT,
    nextMaxed BOOLEAN,
    bank BLOB,
    zoneItems BLOB,
    PRIMARY KEY(user,dt),
    FOREIGN KEY(user) REFERENCES user(id)
);
CREATE TABLE success (
    user INT NOT NULL,
    goal INTEGER NOT NULL,
    accomplished DATETIME(2) NOT NULL,
    amount INT NOT NULL,
    UNIQUE(user,goal,accomplished),
    PRIMARY KEY(user,goal,amount),
    FOREIGN KEY(goal) REFERENCES goal(id),
    FOREIGN KEY(user) REFERENCES user(id)
);
CREATE TABLE invitation (
    user INT NOT NULL,
    challenge INTEGER NOT NULL,
    PRIMARY KEY(user,challenge),
    FOREIGN KEY(challenge) REFERENCES challenge(id)
);
