-- Create database & tables
\benchmark once \name init
DROP SCHEMA IF EXISTS GoBench CASCADE;
CREATE SCHEMA GoBench;
CREATE TABLE GoBench.Customer (CustomerId INT PRIMARY KEY, Name VARCHAR(10), Address VARCHAR(50), Birthday  DATE);
CREATE TABLE "gobench"."order" (OrderId INT PRIMARY KEY, CustomerId INT NOT NULL, CreationDate DATE, Comment VARCHAR(50), FOREIGN KEY (CustomerId) REFERENCES GoBench.Customer (CustomerId));
CREATE TABLE GoBench.Category (CategoryId INT PRIMARY KEY, Name VARCHAR(10));
CREATE TABLE GoBench.Supplier (SupplierId  INT PRIMARY KEY, Name VARCHAR(10), Address VARCHAR(50));
CREATE TABLE GoBench.Product (ProductId INT PRIMARY KEY, SupplierId INT NOT NULL, CategoryId INT NOT NULL, Code VARCHAR(6), Description VARCHAR(100), UnitSize INT, PricePerUnit DECIMAL, FOREIGN KEY (SupplierId) REFERENCES GoBench.Supplier (SupplierId), FOREIGN KEY (CategoryId) REFERENCES GoBench.Category (CategoryId));
CREATE TABLE GoBench.LineItem (LineItemId INT PRIMARY KEY, OrderId INT NOT NULL, ProductId INT NOT NULL, Quantity INT, DeliveryDate DATE, FOREIGN KEY (OrderId) REFERENCES "gobench"."order" (OrderId), FOREIGN KEY (ProductId) REFERENCES GoBench.Product (ProductId));

-- INSERT
\benchmark loop 0.75 \name single
INSERT INTO GoBench.Customer (CustomerId, Name, Address, Birthday) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', '{{call .RandString 10 50 }}', '{{call .RandDate }}');

\benchmark loop 1.0 \name single
INSERT INTO GoBench.Order (OrderId, CustomerId, CreationDate, Comment) VALUES( {{.Iter}}, {{call .RandId "Customer" "SQL" }}, '{{call .RandDate }}', '{{call .RandString 0 50 }}');

\benchmark loop 0.25 \name single
INSERT INTO GoBench.Supplier (SupplierId, Name, Address) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', '{{call .RandString 10 50 }}');

\benchmark loop 0.1 \name single
INSERT INTO GoBench.Category (CategoryId, Name) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}');

\benchmark loop 0.5 \name single
INSERT INTO GoBench.Product (ProductId, SupplierId, CategoryId, Code, Description, UnitSize, PricePerUnit) VALUES( {{.Iter}}, {{call .RandId "Supplier" "SQL" }}, {{call .RandId "Category" "SQL" }}, '{{call .RandString 5 6 }}', '{{call .RandString 0 100 }}', {{call .RandIntBetween 1 10 }}, {{call .RandFloatBetween 0.01 999999.99 }});

\benchmark loop 1.0 \name single
INSERT INTO GoBench.LineItem (LineItemId, OrderId, ProductId, Quantity, DeliveryDate) VALUES( {{.Iter}}, {{call .RandId "Order" "SQL" }}, {{call .RandId "Product" "SQL" }}, {{call .RandIntBetween 1 5000 }}, '{{call .RandDate }}');


-- Delete table
\benchmark once \name clean
DROP TABLE IF EXISTS GoBench.LineItem CASCADE;
DROP TABLE IF EXISTS GoBench.Product CASCADE;
DROP TABLE IF EXISTS GoBench.Category CASCADE;
DROP TABLE IF EXISTS GoBench.Supplier CASCADE;
DROP TABLE IF EXISTS GoBench.Order CASCADE;
DROP TABLE IF EXISTS GoBench.Customer CASCADE;
DROP SCHEMA IF EXISTS GoBench CASCADE;