package code

import (
	"testing"

	"gotest.tools/v3/assert"
)

func Test_GetLOC(t *testing.T) {
	tests := []struct {
		code  string
		lines uint
	}{
		{
			`1111 a aabbaba a a a
	
	
	
	
	22222222  
	          
	
	
	    333333 3 3 33 3 3
      `, 3,
		},
		{
			`1111 a aabbaba a a a
	
	
	
	
	//22222222  
	          
	
	
	    333333 3 3 33 3 3
      `, 2,
		},
		{
			"", 0,
		},
		{
			"  		   ", 0,
		},
		{
			`  		   
			   `, 0,
		},
		{
			`  		   
			  code //comment   `, 1,
		},
		{
			`11 /*
			ss
			ss
			ss
			sss */
			2222`, 2,
		},
		{
			`11 
			/* ss
			ss
			ss
			sss */
			  2222`, 2,
		},
		{
			`11 
			/* ss
			// ss
			ss
			sss */
			  2222`, 2,
		},
		{
			`11 
			/* ss
			 ss
			/* ss
			sss */
			  2222
			  3333333`, 3,
		},
		{
			`11 
			# kdkdkd
			222   22
			3333 #dkdkdkd`, 3,
		},
		{
			`11 
			""" kdkdkd
			222   22
			3333 #dkdkdkd
			jj
			"""    
			222`, 2,
		},
		{
			`11 
			""" kdkdkd
			222   22
			3333 #dkdkdkd
			jj
			dd
			"""    
			222
			3333`, 3,
		},
	}

	for _, tt := range tests {
		res := GetLOC(tt.code)
		assert.Equal(t, res, tt.lines)
	}
}
