-- MySQL Script generated by MySQL Workbench
-- Wed Jan  6 16:06:37 2016
-- Model: New Model    Version: 1.0
-- MySQL Workbench Forward Engineering

SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='TRADITIONAL,ALLOW_INVALID_DATES';

-- -----------------------------------------------------
-- Schema httpDispacher
-- -----------------------------------------------------
DROP SCHEMA IF EXISTS `httpDispacher` ;

-- -----------------------------------------------------
-- Schema httpDispacher
-- -----------------------------------------------------
CREATE SCHEMA IF NOT EXISTS `httpDispacher` DEFAULT CHARACTER SET utf8 ;
SHOW WARNINGS;
USE `httpDispacher` ;

-- -----------------------------------------------------
-- Table `DomainTable`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `DomainTable` ;

SHOW WARNINGS;
CREATE TABLE IF NOT EXISTS `DomainTable` (
  `idDomainName` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `DomainName` CHAR(255) BINARY NOT NULL,
  `NS` VARCHAR(256) NULL,
  PRIMARY KEY (`idDomainName`))
ENGINE = InnoDB;

SHOW WARNINGS;
CREATE UNIQUE INDEX `DomainName_UNIQUE` ON `DomainTable` (`DomainName` ASC);

SHOW WARNINGS;

-- -----------------------------------------------------
-- Table `RegionTable`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `RegionTable` ;

SHOW WARNINGS;
CREATE TABLE IF NOT EXISTS `RegionTable` (
  `idRegion` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `StartIP` INT UNSIGNED ZEROFILL NOT NULL,
  `EndIP` INT UNSIGNED ZEROFILL NOT NULL,
  `NetAddr` INT UNSIGNED NOT NULL,
  `NetMask` INT UNSIGNED NOT NULL,
  PRIMARY KEY (`idRegion`))
ENGINE = InnoDB;

SHOW WARNINGS;
CREATE INDEX `idx_union` USING BTREE ON `RegionTable` (`StartIP` ASC, `EndIP` ASC);

SHOW WARNINGS;

-- -----------------------------------------------------
-- Table `RRTable`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `RRTable` ;

SHOW WARNINGS;
CREATE TABLE IF NOT EXISTS `RRTable` (
  `idRRTable` INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `idDomainName` INT UNSIGNED NOT NULL COMMENT 'id from DomainTable',
  `idRegion` INT UNSIGNED NOT NULL COMMENT 'id From RegionTable',
  `Rrtype` INT UNSIGNED NOT NULL COMMENT '1 for dns.TypeA 5 for dns.CNAME ( golang :github.com/miekg/dns)',
  `Class` INT UNSIGNED NOT NULL DEFAULT 1 COMMENT 'ClassINET   = 1',
  `Ttl` INT UNSIGNED NOT NULL DEFAULT 300,
  `Target` VARCHAR(45) NOT NULL COMMENT 'one IP address or one domain name .',
  PRIMARY KEY (`idRRTable`))
ENGINE = InnoDB;

SHOW WARNINGS;
CREATE INDEX `domain_region_idx` USING BTREE ON `RRTable` (`idDomainName` ASC, `idRegion` ASC);

SHOW WARNINGS;

SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;
