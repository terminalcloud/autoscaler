package main

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var testType = "c3.8xlarge"
var testRam = int64(60000)

var aamock AdminAPI
var apiMock *AdminAPIMock
var as AutoScaler
const test_filter_d = "default"


func prepare() {
	aamock = NewAdminAPIHttpMock()
	apiMock, _ = aamock.(*AdminAPIMock)
	as = NewGeneralAutoScaler(aamock,"default")

	nodeType = &testType
	nodeTypeRamTotal = &testRam

	min_iters_before_shrink = 4
}

func TestStartingAutoScaler(t *testing.T) {

	prepare()

	Convey("Starting the autosclaer", t, func() {
		hns, _, _, err := aamock.GetHNInfo(test_filter_d)
		Convey("Should be empty before running scale", func() {
			So(err, ShouldBeNil)
			So(len(hns), ShouldEqual, 0)
		})
		Convey("Calling Run AutoScale ", func() {
			as.Run()
			apiMock.completeCreates()

			Convey("Should create a node", func() {
				hns, _, _, err := aamock.GetHNInfo(test_filter_d)
				So(err, ShouldBeNil)
				So(len(hns), ShouldEqual, 1)
			})

		})
	})

}

func TestBasicGrow(t *testing.T) {
	prepare()

	Convey("Starting the Autoscaler", t, func() {

		Convey("Calling autoscale", func() {
			as.Run()
			apiMock.completeCreates()
			apiMock.completeDeletes()
			hns,_,_, err := aamock.GetHNInfo(test_filter_d)
			Convey("Should have 1 node", func() {
				So(err, ShouldBeNil)
				So(len(hns), ShouldEqual, 1)
			})
		})

		Convey("Increase cluster load  > 0.8", func() {
			hns, _, _, _ := aamock.GetHNInfo(test_filter_d)
			apiMock.modifyLoad(hns[0].NodeIp, int64(59000))
			fmt.Println("Running after load increase")
			as.Run()
			apiMock.completeCreates()
			apiMock.completeDeletes()
			fmt.Println(hns[0])

			Convey("Should create 2nd node", func() {
				hns, _, _, err := apiMock.GetHNInfo(test_filter_d)
				So(err, ShouldBeNil)
				So(len(hns), ShouldEqual, 2)
			})

		})
	})
}

func TestBufferZone(t *testing.T) {
	prepare()
	Convey("Starting the Autoscaler", t, func() {
		Convey("Test cluster load to be  0.7<l<0.8", func() {

			Convey("Should initially be empty", func() {
				hns, _, _, err := aamock.GetHNInfo(test_filter_d)
				So(err, ShouldBeNil)
				So(len(hns), ShouldEqual, 0)

				Convey("Changing load on cluster", func() {
					apiMock.populateHealth(int64(45000), 2)
					as.Run()
					apiMock.completeCreates()
					apiMock.completeDeletes()

					Convey("Should create 2nd node", func() {
						hns, _, _, err := aamock.GetHNInfo(test_filter_d)
						So(err, ShouldBeNil)
						So(len(hns), ShouldEqual, 2)
					})

				})
			})
		})
	})
}

func TestShrink(t *testing.T) {
	prepare()
	Convey("Starting the Autoscaler", t, func() {

		Convey("Testing cluster shrink conditions", func() {

			Convey("Should initially be empty", func() {
				hns, _, _, err := aamock.GetHNInfo(test_filter_d)
				So(err, ShouldBeNil)
				So(len(hns), ShouldEqual, 0)

				Convey("Changing load on cluster to 0.7 pre shrink", func() {
					apiMock.populateHealth(int64(25000), 3)

					for i := 0; i <= min_iters_before_shrink; i++ {
						as.Run()
						apiMock.completeCreates()
						apiMock.completeDeletes()
					}

					Convey("Should Now have just 2 nodes", func() {
						hns, _, _, err := aamock.GetHNInfo(test_filter_d)
						So(err, ShouldBeNil)
						So(len(hns), ShouldEqual, 2)

						Convey("Should not delete the last node as it will have to grow", func() {
							as.Run()
							apiMock.completeCreates()
							apiMock.completeDeletes()

							Convey("Should still have 2 nodes left", func() {
								hns, _, _, err := aamock.GetHNInfo(test_filter_d)
								fmt.Println(hns[0])
								So(err, ShouldBeNil)
								So(len(hns), ShouldEqual, 2)
							})
						})
					})
				})
			})
		})
	})
}

/* 
 * Fix Test by creating API to simulate pending in HN Health
func TestBuffer(t *testing.T) {
	prepare()

	Convey("Starting the Autoscaler", t, func() {

		Convey("Should initially be empty", func() {
			hns, _, _, err := aamock.GetHNInfo(test_filter_d)
			So(err, ShouldBeNil)
			So(len(hns), ShouldEqual, 0)

			Convey("Empty should trigger node creation", func() {
				as.Run()
				So(len(apiMock.internalReq), ShouldEqual, 1)
			})

			Convey("Running again should not create new node", func() {
				as.Run()
				So(len(apiMock.internalReq), ShouldEqual, 1)
			})

			Convey("Complete create", func() {
				apiMock.completeCreates()
				hns, _, _, err := aamock.GetHNInfo(test_filter_d)
				So(err, ShouldBeNil)
				So(len(hns), ShouldEqual, 1)
			})
		})
	})
}
*/