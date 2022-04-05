-- INIT
\benchmark once \name initialize
DROP DATABASE IF EXISTS GoBench; CREATE DATABASE GoBench; USE GoBench; CREATE TABLE Customer (CustomerId INT PRIMARY KEY, Name VARCHAR(10), Address VARCHAR(50), Birthday  DATE);    CREATE TABLE `Order` (OrderId INT PRIMARY KEY, CustomerId INT NOT NULL, CreationDate DATE, Comment VARCHAR(50), FOREIGN KEY (CustomerId) REFERENCES Customer (CustomerId));    CREATE TABLE Category (CategoryId INT PRIMARY KEY, Name VARCHAR(10));    CREATE TABLE Supplier (SupplierId  INT PRIMARY KEY, Name VARCHAR(10), Address VARCHAR(50));    CREATE TABLE Product (ProductId INT PRIMARY KEY, SupplierId INT NOT NULL, CategoryId INT NOT NULL, Code VARCHAR(6), Description VARCHAR(100), UnitSize INT, PricePerUnit DECIMAL(10,2), FOREIGN KEY (SupplierId) REFERENCES Supplier (SupplierId), FOREIGN KEY (CategoryId) REFERENCES Category (CategoryId));    CREATE TABLE LineItem (LineItemId INT PRIMARY KEY, OrderId INT NOT NULL, ProductId INT NOT NULL, Quantity INT, DeliveryDate DATE, FOREIGN KEY (OrderId) REFERENCES `Order` (OrderId), FOREIGN KEY (ProductId) REFERENCES Product (ProductId));

-- INSERT
\benchmark loop 1.0 \name insert_customer
INSERT INTO GoBench.Customer (CustomerId, Name, Address, Birthday) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', '{{call .RandString 10 50 }}', '{{call .RandDate }}');

\benchmark loop 1.0 \name insert_order
INSERT INTO GoBench.Order (OrderId, CustomerId, CreationDate, Comment) VALUES( {{.Iter}}, {{.Iter}}, '{{call .RandDate }}', '{{call .RandString 0 50 }}');

\benchmark loop 1.0 \name insert_supplier
INSERT INTO GoBench.Supplier (SupplierId, Name, Address) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', '{{call .RandString 10 50 }}');

\benchmark loop 1.0 \name insert_category
INSERT INTO GoBench.Category (CategoryId, Name) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}');

\benchmark loop 1.0 \name insert_product
INSERT INTO GoBench.Product (ProductId, SupplierId, CategoryId, Code, Description, UnitSize, PricePerUnit) VALUES( {{.Iter}}, {{.Iter}}, {{.Iter}}, '{{call .RandString 5 6 }}', '{{call .RandString 0 100 }}', {{call .RandIntBetween 1 10 }}, {{call .RandFloatBetween 0.01 999999.99 }});

\benchmark loop 1.0 \name insert_lineitem
INSERT INTO GoBench.LineItem (LineItemId, OrderId, ProductId, Quantity, DeliveryDate) VALUES( {{.Iter}}, {{.Iter}}, {{.Iter}}, {{call .RandIntBetween 1 5000 }}, '{{call .RandDate }}');


-- SELECTS
\benchmark loop 1.0 \name select_simple
SELECT * FROM GoBench.Customer WHERE CustomerId = {{.Iter}}

\benchmark loop 1.0 \name select_medium
SELECT * FROM GoBench.Product p JOIN GoBench.Supplier s on p.SupplierId = s.SupplierId WHERE s.SupplierId = {{.Iter}} ORDER BY p.PricePerUnit DESC

\benchmark loop 1.0 \name select_complex
SELECT c.CustomerId, c.Name, SUM(li.Quantity * p.UnitSize  * p.PricePerUnit) as TotalOrderValue  FROM GoBench.Customer c  INNER JOIN GoBench.Order o on o.CustomerId = c.CustomerId  INNER JOIN GoBench.LineItem li on o.OrderId = li.OrderId  INNER JOIN GoBench.Product p on p.ProductId = li.ProductId  WHERE (o.CreationDate BETWEEN '{{call .RandDate }}' AND '9999-12-31')  GROUP by c.CustomerId, c.Name  ORDER by c.CustomerId




-- CLEAN
--\benchmark once \name clean
--USE GoBench; SET FOREIGN_KEY_CHECKS=0; DROP TABLE GoBench.LineItem; DROP TABLE GoBench.Order; DROP TABLE GoBench.Customer; DROP TABLE GoBench.Product; DROP TABLE GoBench.Category; DROP TABLE GoBench.Supplier; SET FOREIGN_KEY_CHECKS=1; DROP DATABASE GoBench;