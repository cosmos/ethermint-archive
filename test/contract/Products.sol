contract Products {
    struct Product {
        uint id;
        string name;
        address manufacturer;
    }

    event NewProduct (
        uint id,
        address indexed manufacturer,
        string name
    );

    mapping(uint => Product) products;
    mapping(address => uint[]) productsOfManufacturer;

    function registerProduct(uint id, string name) {
        products[id] = Product(id, name, msg.sender);
        productsOfManufacturer[msg.sender].push(id);
        NewProduct(products[id].id, products[id].manufacturer, products[id].name);
    }

    function getProduct(uint id) constant returns(uint,string, address) {
        var product = products[id];
        return (product.id, product.name, products[id].manufacturer);
    }
    
    function getProducts(address manufacturer) constant returns(uint[]) {
        return productsOfManufacturer[manufacturer];
    }
}
