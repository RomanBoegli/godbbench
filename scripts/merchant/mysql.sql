-- INIT
\benchmark once \name initialize
DROP DATABASE IF EXISTS godbbench; 
CREATE DATABASE godbbench; 
USE godbbench; 
CREATE TABLE Customer (CustomerId INT PRIMARY KEY, Name VARCHAR(10), Address VARCHAR(50), Birthday  DATE);
CREATE TABLE `Order` (OrderId INT PRIMARY KEY, CustomerId INT NOT NULL, CreationDate DATE, Comment VARCHAR(50), FOREIGN KEY (CustomerId) REFERENCES Customer (CustomerId));
CREATE TABLE Category (CategoryId INT PRIMARY KEY, Name VARCHAR(10));
CREATE TABLE Supplier (SupplierId  INT PRIMARY KEY, Name VARCHAR(10), Address VARCHAR(50));
CREATE TABLE Product (ProductId INT PRIMARY KEY, SupplierId INT NOT NULL, CategoryId INT NOT NULL, Code VARCHAR(6), Description VARCHAR(100), UnitSize INT, PricePerUnit DECIMAL(10,2), FOREIGN KEY (SupplierId) REFERENCES Supplier (SupplierId), FOREIGN KEY (CategoryId) REFERENCES Category (CategoryId)); 
CREATE TABLE LineItem (LineItemId INT PRIMARY KEY, OrderId INT NOT NULL, ProductId INT NOT NULL, Quantity INT, DeliveryDate DATE, FOREIGN KEY (OrderId) REFERENCES `Order` (OrderId), FOREIGN KEY (ProductId) REFERENCES Product (ProductId));

-- INSERTS
\benchmark loop 1.00 \name inserts
INSERT INTO godbbench.Customer (CustomerId, Name, Address, Birthday) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', '{{call .RandString 10 50 }}', '{{call .RandDate }}');
INSERT INTO godbbench.Order (OrderId, CustomerId, CreationDate, Comment) VALUES( {{.Iter}}, (SELECT CustomerId FROM godbbench.Customer ORDER BY RAND() LIMIT 1), '{{call .RandDate }}', '{{call .RandString 0 50 }}');
INSERT INTO godbbench.Supplier (SupplierId, Name, Address) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', '{{call .RandString 10 50 }}');
INSERT INTO godbbench.Category (CategoryId, Name) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}');
INSERT INTO godbbench.Product (ProductId, SupplierId, CategoryId, Code, Description, UnitSize, PricePerUnit) VALUES( {{.Iter}}, (SELECT SupplierId FROM godbbench.Supplier ORDER BY RAND() LIMIT 1), (SELECT CategoryId FROM godbbench.Category ORDER BY RAND() LIMIT 1), '{{call .RandString 5 6 }}', '{{call .RandString 0 100 }}', {{call .RandIntBetween 1 10 }}, {{call .RandFloatBetween 0.01 999999.99 }});
INSERT INTO godbbench.LineItem (LineItemId, OrderId, ProductId, Quantity, DeliveryDate) VALUES( {{.Iter}}, (SELECT OrderId FROM godbbench.Order ORDER BY RAND() LIMIT 1), (SELECT ProductId FROM godbbench.Product ORDER BY RAND() LIMIT 1), {{call .RandIntBetween 1 5000 }}, '{{call .RandDate }}');

-- SELECTS
\benchmark loop 1.0 \name select_simple
SELECT * FROM godbbench.Customer WHERE CustomerId = {{.Iter}};

\benchmark loop 1.0 \name select_medium
SELECT * FROM godbbench.Product p JOIN godbbench.Supplier s on p.SupplierId = s.SupplierId WHERE s.SupplierId = {{.Iter}} ORDER BY p.PricePerUnit DESC;

\benchmark loop 1.0 \name select_complex
SELECT c.CustomerId, c.Name, SUM(li.Quantity * p.UnitSize  * p.PricePerUnit) as TotalOrderValue  FROM godbbench.Customer c  INNER JOIN godbbench.Order o on o.CustomerId = c.CustomerId  INNER JOIN godbbench.LineItem li on o.OrderId = li.OrderId  INNER JOIN godbbench.Product p on p.ProductId = li.ProductId  WHERE (o.CreationDate BETWEEN '{{call .RandDate }}' AND '9999-12-31')  GROUP by c.CustomerId, c.Name  ORDER by c.CustomerId;

-- CLEAN
\benchmark once \name clean
USE godbbench; 
SET FOREIGN_KEY_CHECKS=0;
DROP DATABASE godbbench;
SET FOREIGN_KEY_CHECKS=1;
