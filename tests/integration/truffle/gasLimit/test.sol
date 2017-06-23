pragma solidity ^0.4.0;
contract Test
{
    event TestEvent(uint a);

    function Test() {
    }

    function testF1(uint i)
    {
        TestEvent(i);
    }

    function testF2(uint i)
    {
        TestEvent(i);
    }

    function testF3(uint i)
    {
        TestEvent(i);
    }

    function testF4(uint i)
    {
        TestEvent(i);
    }
}